package cronjob

import (
	"strconv"
	"strings"
	"time"
)

type CommandImp interface {
	//Name 名称
	Name() string
	//Func 方法
	Func() func()
	//Expression cron 表达式: "* * * * * *"
	Expression() string
	//Cron 自定义 expression "* * * * * *"
	Cron(expression string) CommandImp
	//EverySecond 每 1 秒
	EverySecond() CommandImp
	//EveryTwoSeconds 每 2 秒
	EveryTwoSeconds() CommandImp
	//EveryThreeSeconds 每 3 秒
	EveryThreeSeconds() CommandImp
	//EveryFourSeconds 每 4 秒
	EveryFourSeconds() CommandImp
	//EveryFiveSeconds 每 5 秒
	EveryFiveSeconds() CommandImp
	//EveryTenSeconds 每 10 秒
	EveryTenSeconds() CommandImp
	//EveryFifteenSeconds 每 15 秒
	EveryFifteenSeconds() CommandImp
	//EveryThirtySeconds 每 30 秒
	EveryThirtySeconds() CommandImp
	//EveryMinute 每分钟
	EveryMinute() CommandImp
	//EveryTwoMinutes 每 2 分钟
	EveryTwoMinutes() CommandImp
	//EveryThreeMinutes 每 3 分钟
	EveryThreeMinutes() CommandImp
	//EveryFourMinutes 每 4 分钟
	EveryFourMinutes() CommandImp
	//EveryFiveMinutes 每 5 分钟
	EveryFiveMinutes() CommandImp
	//EveryTenMinutes 每 10 分钟
	EveryTenMinutes() CommandImp
	//EveryFifteenMinutes 每 15 分钟
	EveryFifteenMinutes() CommandImp
	//EveryThirtyMinutes 每 30 分钟
	EveryThirtyMinutes() CommandImp
	//Hourly 每小时
	Hourly() CommandImp
	//HourlyAt 每小时的第几分钟
	HourlyAt([]int) CommandImp
	//EveryTwoHours 每 2 小时
	EveryTwoHours() CommandImp
	//EveryThreeHours 每 3 小时
	EveryThreeHours() CommandImp
	//EveryFourHours 每 4 小时
	EveryFourHours() CommandImp
	//EverySixHours 每 6 小时
	EverySixHours() CommandImp
	//Daily 每天
	Daily() CommandImp
	//DailyAt 每天几点(time: "2:00")
	DailyAt(time string) CommandImp
	//At alias of DailyAt
	At(string) CommandImp
	//Weekdays 工作日 1-5
	Weekdays() CommandImp
	//Weekends 周末
	Weekends() CommandImp
	//Mondays 周一
	Mondays() CommandImp
	//Tuesdays 周二
	Tuesdays() CommandImp
	//Wednesdays 周三
	Wednesdays() CommandImp
	//Thursdays 周四
	Thursdays() CommandImp
	//Fridays 周五
	Fridays() CommandImp
	//Saturdays 周六
	Saturdays() CommandImp
	//Sundays 周日
	Sundays() CommandImp
	//Weekly 每周一
	Weekly() CommandImp
	//WeeklyOn 周日几(day) 几点(time: "0:0")
	WeeklyOn(day int, time string) CommandImp
	Monthly() CommandImp
	// MonthlyOn dayOfMonth: 1, time: "0:0"
	MonthlyOn(dayOfMonth string, time string) CommandImp
	//LastDayOfMonth 每月最后一天
	LastDayOfMonth(time string) CommandImp
	// Quarterly 每季度执行
	Quarterly() CommandImp
	// QuarterlyOn 每季度的第几天，几点(time: "0:0")执行
	QuarterlyOn(dayOfQuarter string, time string) CommandImp
	//Yearly 每年
	Yearly() CommandImp
	//YearlyOn 每年几月(month) 哪天(dayOfMonth) 时间(time: "0:0")
	YearlyOn(month string, dayOfMonth string, time string) CommandImp
	//Days 天(0-6: 周日-周六)
	Days([]int) CommandImp
}

// second minute hour `day of the month` month `day of the week`
const expression = "* * * * * *"

const (
	pos_second = iota
	pos_minute
	pos_hour
	pos_day_of_month
	pos_month
	pos_day_of_week
)

const (
	sunday = iota
	monday
	tuesday
	wednesday
	thursday
	friday
	saturday
)

type command struct {
	name       string
	expression string

	fn func()
}

type onceCommand struct {
	name string
	id   int
	date time.Time

	fn func()
}

func (c *command) Func() func() {
	return c.fn
}

func (c *command) Expression() string {
	return c.expression
}

func (c *command) Name() string {
	return c.name
}

func (c *command) Cron(expression string) CommandImp {
	c.expression = expression
	return c
}

func (c *command) EverySecond() CommandImp {
	c.spliceIntoPosition(pos_second, "*")
	return c
}

func (c *command) EveryTwoSeconds() CommandImp {
	c.spliceIntoPosition(pos_second, "*/2")
	return c
}

func (c *command) EveryThreeSeconds() CommandImp {
	c.spliceIntoPosition(pos_second, "*/3")
	return c
}

func (c *command) EveryFourSeconds() CommandImp {
	c.spliceIntoPosition(pos_second, "*/4")
	return c
}

func (c *command) EveryFiveSeconds() CommandImp {
	c.spliceIntoPosition(pos_second, "*/5")
	return c
}

func (c *command) EveryTenSeconds() CommandImp {
	c.spliceIntoPosition(pos_second, "*/10")
	return c
}

func (c *command) EveryFifteenSeconds() CommandImp {
	c.spliceIntoPosition(pos_second, "*/15")
	return c
}

func (c *command) EveryThirtySeconds() CommandImp {
	c.spliceIntoPosition(pos_second, "0,30")
	return c
}

func (c *command) EveryMinute() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "*")
	return c
}

func (c *command) EveryTwoMinutes() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "*/2")
	return c
}

func (c *command) EveryThreeMinutes() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "*/3")
	return c
}

func (c *command) EveryFourMinutes() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "*/4")
	return c
}

func (c *command) EveryFiveMinutes() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "*/5")
	return c
}

func (c *command) EveryTenMinutes() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "*/10")
	return c
}

func (c *command) EveryFifteenMinutes() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "*/15")
	return c
}

func (c *command) EveryThirtyMinutes() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "0,30")
	return c
}

func (c *command) Hourly() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "0")
	return c
}

func (c *command) HourlyAt(minutes []int) CommandImp {
	var minsStr []string
	for _, day := range minutes {
		minsStr = append(minsStr, strconv.Itoa(day))
	}
	if len(minutes) == 0 {
		minsStr = []string{"0"}
	}
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, strings.Join(minsStr, ","))
	return c
}

func (c *command) EveryTwoHours() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "0")
	c.spliceIntoPosition(pos_hour, "*/2")
	return c
}

func (c *command) EveryThreeHours() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "0")
	c.spliceIntoPosition(pos_hour, "*/3")
	return c
}

func (c *command) EveryFourHours() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "0")
	c.spliceIntoPosition(pos_hour, "*/4")
	return c
}

func (c *command) EverySixHours() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "0")
	c.spliceIntoPosition(pos_hour, "*/6")
	return c
}

func (c *command) Daily() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "0")
	c.spliceIntoPosition(pos_hour, "0")
	return c
}

func (c *command) At(time string) CommandImp {
	return c.DailyAt(time)
}

func (c *command) DailyAt(time string) CommandImp {
	hour, minute := "0", "0"
	if time != "" {
		split := strings.Split(time, ":")
		if len(split) == 2 {
			minute = split[1]
		}
		hour = split[0]
	}
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_hour, hour)
	c.spliceIntoPosition(pos_minute, minute)
	return c
}

func (c *command) Weekdays() CommandImp {
	return c.Days([]int{monday, tuesday, wednesday, thursday, friday})
}

func (c *command) Weekends() CommandImp {
	c.Days([]int{saturday, sunday})
	return c
}

func (c *command) Mondays() CommandImp {
	c.Days([]int{monday})
	return c
}

func (c *command) Tuesdays() CommandImp {
	c.Days([]int{tuesday})
	return c
}

func (c *command) Wednesdays() CommandImp {
	c.Days([]int{wednesday})
	return c
}

func (c *command) Thursdays() CommandImp {
	c.Days([]int{thursday})
	return c
}

func (c *command) Fridays() CommandImp {
	c.Days([]int{friday})
	return c
}

func (c *command) Saturdays() CommandImp {
	c.Days([]int{saturday})
	return c
}

func (c *command) Sundays() CommandImp {
	c.Days([]int{sunday})
	return c
}

func (c *command) Weekly() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "0")
	c.spliceIntoPosition(pos_hour, "0")
	c.spliceIntoPosition(pos_day_of_week, "1")
	return c
}

func (c *command) WeeklyOn(day int, time string) CommandImp {
	if time == "" {
		time = "0:0"
	}
	c.DailyAt(time)
	c.Days([]int{day})
	return c
}

func (c *command) Monthly() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "0")
	c.spliceIntoPosition(pos_hour, "0")
	c.spliceIntoPosition(pos_day_of_month, "1")
	return c
}

func (c *command) MonthlyOn(dayOfMonth string, time string) CommandImp {
	if dayOfMonth == "" {
		dayOfMonth = "1"
	}
	if time == "" {
		time = "0:0"
	}
	c.DailyAt(time)
	c.spliceIntoPosition(pos_day_of_month, dayOfMonth)
	return c
}

func (c *command) LastDayOfMonth(time string) CommandImp {
	c.DailyAt(time)
	c.spliceIntoPosition(pos_day_of_month, "L")
	return c
}

func (c *command) Quarterly() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "0")
	c.spliceIntoPosition(pos_hour, "0")
	c.spliceIntoPosition(pos_day_of_month, "1")
	c.spliceIntoPosition(pos_month, "1-12/3")
	return c
}

func (c *command) QuarterlyOn(dayOfQuarter string, time string) CommandImp {
	if dayOfQuarter == "" {
		dayOfQuarter = "1"
	}
	c.DailyAt(time)
	c.spliceIntoPosition(pos_day_of_month, dayOfQuarter)
	c.spliceIntoPosition(pos_month, "1-12/3")
	return c
}

func (c *command) Yearly() CommandImp {
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_minute, "0")
	c.spliceIntoPosition(pos_hour, "0")
	c.spliceIntoPosition(pos_day_of_month, "1")
	c.spliceIntoPosition(pos_month, "1")
	return c
}

func (c *command) YearlyOn(month string, dayOfMonth string, time string) CommandImp {
	c.DailyAt(time)
	c.spliceIntoPosition(pos_day_of_month, dayOfMonth)
	c.spliceIntoPosition(pos_month, month)
	return c
}

func (c *command) Days(days []int) CommandImp {
	var daysStr []string
	for _, day := range days {
		daysStr = append(daysStr, strconv.Itoa(day))
	}
	c.spliceIntoPosition(pos_second, "0")
	c.spliceIntoPosition(pos_day_of_week, strings.Join(daysStr, ","))
	return c
}

func (c *command) spliceIntoPosition(pos int, val string) {
	split := strings.Split(c.expression, " ")
	split[pos] = val
	c.expression = strings.Join(split, " ")
}
