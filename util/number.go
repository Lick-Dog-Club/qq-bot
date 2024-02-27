package util

import "strconv"

func ToFloat64(s string) float64 {
	float, _ := strconv.ParseFloat(s, 64)
	return float
}
func ToInt64(s string) int64 {
	ints, _ := strconv.ParseInt(s, 10, 64)
	return ints
}
