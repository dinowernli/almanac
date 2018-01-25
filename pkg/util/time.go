package util

import "time"

const (
	nanosPerMilli = 1000000
)

func TimeMs(millis int64) time.Time {
	return time.Unix(0, millis*nanosPerMilli)
}
