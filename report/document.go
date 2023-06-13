package main

import (
	"encoding/json"
	"errors"
	"regexp"
	"time"
)

const (
	KEY_INDEX     = "_index"
	KEY_ID        = "_id"
	KEY_TIMESTAMP = "@timestamp"
	KEY_HOST      = "host"
	KEY_UUID      = "uuid"
	KEY_KIND      = "kind"
	kEY_COUNTRY   = "geoip_country_name"
	kEY_CITY      = "geoip_region_name"
	KEY_USAGE     = "usage"
)

var (
	ErrInvalidUUID      = errors.New("invalid message uuid")
	ErrInvalidTimestamp = errors.New("invalid message timestamp")
)

type document struct {
	index     string        // "curvebs_usage_report"
	id        string        // "876A5499646189AB0FB0D21EB9CFB871"
	timestamp string        // "2023-05-26T07:00:01.992Z"
	host      string        // "60.191.87.80"
	uuid      string        // "876A5499646189AB0FB0D21EB9CFB871"
	kind      string        // curvebs
	country   []interface{} // ["China","China]"
	city      []interface{} // ["Guangdong","Guangdong"]
	usage     float64       // 12345

	date time.Time
}

func strfield(field interface{}) string {
	if field == nil {
		return ""
	}
	return field.(string)
}

func numfield(field interface{}) float64 {
	if field == nil {
		return 0
	}
	return field.(float64)
}

func slicefield(field interface{}) []interface{} {
	if field == nil {
		return []interface{}{}
	} else if _, ok := field.([]interface{}); !ok {
		return []interface{}{}
	}

	return field.([]interface{})
}

func (doc *document) check() error {
	if len(doc.uuid) == 0 {
		return ErrInvalidUUID
	} else if len(doc.timestamp) == 0 {
		return ErrInvalidTimestamp
	}

	pattern := regexp.MustCompile("^([0-9]{4}-[0-9]{2}-[0-9]{2})(.+)$")
	mu := pattern.FindStringSubmatch(doc.timestamp)
	if len(mu) == 0 {
		return ErrInvalidTimestamp
	}
	date, err := time.Parse("2006-01-02", mu[1])
	if err != nil {
		return ErrInvalidTimestamp
	}
	doc.date = date

	return nil
}

func (doc *document) decode(data string) error {
	m := map[string]interface{}{}
	err := json.Unmarshal([]byte(data), &m)
	if err != nil {
		return err
	}

	doc.index = strfield(m[KEY_INDEX])
	doc.id = strfield(m[KEY_ID])
	doc.timestamp = strfield(m[KEY_TIMESTAMP])
	doc.host = strfield(m[KEY_HOST])
	doc.uuid = strfield(m[KEY_UUID])
	doc.kind = strfield(m[KEY_KIND])
	doc.usage = numfield(m[KEY_USAGE])
	doc.country = slicefield(m[kEY_COUNTRY])
	doc.city = slicefield(m[kEY_CITY])
	return doc.check()
}

func (doc *document) encode() (string, error) {
	bytes, err := json.Marshal(*doc)
	return string(bytes), err
}
