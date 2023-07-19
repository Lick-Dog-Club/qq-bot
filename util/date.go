package util

import "time"

func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
}

func Yesterday() time.Time {
	return Today().Add(-24 * time.Hour)
}

func Tomorrow() time.Time {
	return Today().Add(24 * time.Hour)
}
