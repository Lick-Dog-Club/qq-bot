package holiday

import (
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	n := time.Date(2023, 12, 31, 0, 0, 0, 0, time.Local)
	res := getNextHolidays(2023, n)
	if len(res) < 1 {
		res = getNextHolidays(2023+1, n)
	}
	t.Log(res)
}
