package util

import "strconv"

func ToFloat64(s string) float64 {
	float, _ := strconv.ParseFloat(s, 64)
	return float
}
