package main

import (
	"log"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type config struct {
	RedisAddr string `key:"redis_addr" default:"127.0.0.1:6379"`
	Index     string `key:"elastic_index" default:"alive_daily"`
	Version   string `key:"elastic_version" default:"beta"`

	StartWorker   string `key:"start_worker" default:"true"`
	StartReceiver string `key:"start_receiver" default:"true"`
	StartDay      string `key:"start_day" default:"2022-01-01"`
	Debug         string `key:"debug" default:"false"`
}

type reporter struct {
	cfg      config
	rdb      *redis.Client
	clusters map[string]cluster
	start    time.Time
	handling time.Time
	wg       sync.WaitGroup
}

func parseCfg() config {
	cfg := config{}
	t := reflect.TypeOf(cfg)
	v := reflect.ValueOf(&cfg)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		realValue := v.Elem().FieldByName(field.Name)
		key := field.Tag.Get("key")
		value := os.Getenv(key)
		if len(value) == 0 {
			value = field.Tag.Get("default")
		}

		realValue.SetString(value)
		log.Printf("[info] %s => %s\n", key, value)
	}
	return cfg
}

func create() *reporter {
	cfg := parseCfg()
	return &reporter{
		cfg: cfg,
		rdb: redis.NewClient(&redis.Options{
			Addr: cfg.RedisAddr,
			DB:   0, // use default DB
		}),
		clusters: map[string]cluster{},
	}
}

func (r *reporter) init() error {
	var err error
	r.start, err = parse(r.cfg.StartDay)
	return err
}

func (r *reporter) run() {
	if r.cfg.StartReceiver == "true" {
		go r.receiver()
	}
	if r.cfg.StartWorker == "true" {
		go r.worker()
	}
	r.wg.Add(2)
}

func (r *reporter) wait() {
	r.wg.Wait()
}

func (r *reporter) debug(format string, v ...any) {
	if r.cfg.Debug == "false" {
		return
	}
	log.Printf(format, v...)
}
