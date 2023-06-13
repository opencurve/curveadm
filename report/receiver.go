package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	LIST_ELASTIC_TO_REDIS = "elastic-to-redis"
)

func (r *reporter) llen2redis() int64 {
	cmd := r.rdb.LLen(context.Background(), LIST_ELASTIC_TO_REDIS)
	len, err := cmd.Result()
	if err != nil {
		log.Printf("[error] llen %s: %v", LIST_ELASTIC_TO_REDIS, err)
		return 0
	}
	return len
}

func (r *reporter) lpop2redis() (string, error) {
	cmd := r.rdb.LPop(context.Background(), LIST_ELASTIC_TO_REDIS)
	message, err := cmd.Result()
	if err != nil {
		return "", fmt.Errorf("lpop %s: %v", LIST_ELASTIC_TO_REDIS, err)
	}
	return message, nil
}

func (r *reporter) hset2redis(key, field, value string) error {
	cmd := r.rdb.HSet(context.Background(), key, field, value)
	err := cmd.Err()
	if err == nil {
		r.debug("[info] receiver => HSET %s %s", key, field)
		r.debug("[info] receiver => %s", value)
	}
	return err
}

// HSET key field value
// HSET alive-daily-2023-06-12 876A5499646189AB0FB0D21EB9CFB871 {"uuid":...}
func (r *reporter) receive_() error {
	message, err := r.lpop2redis()
	if err != nil {
		return err
	}
	r.debug("[info] receiver <= %s", message)

	var doc document
	err = doc.decode(message)
	if err != nil {
		return err
	}

	key := hashname(strday(doc.date))
	return r.hset2redis(key, doc.uuid, message)
}

func (r *reporter) once(n int64) {
	log.Printf("[info] receive %d messages, handling...\n", n)

	var wg sync.WaitGroup
	workers := make(chan struct{}, 100)
	for atomic.LoadInt64(&n) > 0 {
		workers <- struct{}{}
		wg.Add(1)
		go func() {
			err := r.receive_()
			if err != nil {
				log.Printf("[error] handle err = %v", err)
			}
			<-workers
			wg.Done()
		}()
		atomic.AddInt64(&n, -1)
	}
	wg.Wait()
}

// receive message from redis list and restore to redis hash table.
func (r *reporter) receiver() {
	for {
		n := r.llen2redis()
		if n == 0 {
			time.Sleep(time.Duration(3) * time.Second)
			continue
		}
		r.once(n)
	}
}
