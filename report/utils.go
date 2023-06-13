package main

import (
	"fmt"
	"time"
)

const (
	FORMAT_HASH_ALIVE_DAILY = "alive-daily-%s"
)

func strday(day time.Time) string {
	return day.Format("2006-01-02")
}

// alive-daily-2023-06-13
func hashname(day string) string {
	return fmt.Sprintf(FORMAT_HASH_ALIVE_DAILY, day)
}

func parse(day string) (time.Time, error) {
	return time.Parse("2006-01-02", day)
}
