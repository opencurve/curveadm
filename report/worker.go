package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

const (
	LIST_REDIS_TO_ELASTIC = "redis-to-elastic"
)

type output struct {
	Index     string `json:"index"`      // "alive_daily"
	Id        string `json:"id"`         // "876A5499646189AB0FB0D21EB9CFB871"
	Timestamp string `json:"@timestamp"` // "2023-05-26T07:00:01.992Z"
	Version   string `json:"version"`    // "v0.0.1"

	Host    string `json:"host"`    // "60.191.87.80"
	Uuid    string `json:"uuid"`    // "876A5499646189AB0FB0D21EB9CFB871"
	Kind    string `json:"kind"`    // "curvebs"
	Running int    `json:"running"` // 10
	Alive   int    `json:"alive"`   // 14
	Birth   string `json:"birth"`   // 2023-02-06

	Country []interface{} `json:"geoip_country_name"` // "["China","China]"
	City    []interface{} `json:"geoip_region_name"`  // ["Guangdong","Guangdong"]
	Usage   float64       `json:"usage"`              // 12345
}

func (out *output) encode() (string, error) {
	bytes, err := json.Marshal(*out)
	return string(bytes), err
}

type cluster struct {
	birth time.Time
	alive map[time.Time]bool
}

func (r *reporter) heartbeat(uuid string, day time.Time) {
	cls, ok := r.clusters[uuid]
	if !ok {
		cls = cluster{birth: day, alive: map[time.Time]bool{}}
	}
	cls.alive[day] = true
	r.clusters[uuid] = cls
}

func (r *reporter) cluster(uuid string) cluster {
	return r.clusters[uuid]
}

func (r *reporter) lpush2redis(out *output) error {
	message, err := out.encode()
	if err != nil {
		return err
	}

	r.debug("[info] worker => %s", message)
	cmd := r.rdb.LPush(context.Background(), LIST_REDIS_TO_ELASTIC, message)
	return cmd.Err()
}

func (r *reporter) handle_(message string) error {
	r.debug("[info] worker <= %s", message)

	// 1) decode to document
	var doc document
	err := doc.decode(message)
	if err != nil {
		return err
	}

	// 2) update cluster status
	r.heartbeat(doc.uuid, doc.date)

	// 3) generate output document
	var out output
	cluster := r.cluster(doc.uuid)
	out.Index = r.cfg.Index
	out.Id = fmt.Sprintf("%s-%s", doc.uuid, strday(doc.date))
	out.Timestamp = doc.timestamp
	out.Host = doc.host
	out.Uuid = doc.uuid
	out.Kind = doc.kind
	out.Running = len(cluster.alive)
	out.Alive = int((doc.date.Sub(cluster.birth).Hours() + 24) / 24)
	out.Birth = strday(cluster.birth)
	out.Version = r.cfg.Version
	out.Usage = doc.usage
	out.Country = doc.country
	out.City = doc.city
	return r.lpush2redis(&out)
}

func (r *reporter) handle(day time.Time) error {
	r.debug("[info] worker ==========> %s", strday(day))

	key := hashname(strday(day))
	cmd := r.rdb.HGetAll(context.Background(), key)
	messages, err := cmd.Result()
	if err != nil {
		return err
	} else if len(messages) == 0 {
		return nil
	}

	for _, message := range messages {
		err := r.handle_(message)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *reporter) traverse(handling time.Time) {
	d := handling
	end := time.Now()
	for {
		if d.After(end) {
			break
		}

		err := r.handle(d)
		if err != nil {
			log.Printf("[error] handle err: %v", err)
			return
		}

		r.handling = d
		d = d.AddDate(0, 0, 1)
	}
}

// read all messages from redis hash table and handle it,
// then restore to redis list which will sended to elasticsearch by logstash.
func (r *reporter) worker() {
	r.handling = r.start
	for {
		r.traverse(r.handling)
		time.Sleep(time.Duration(1) * time.Hour)
	}
}
