package etc

import (
	"fmt"
	"time"
)

func GetTodayStartDateTime() string {
	now := time.Now().UTC()
	nowJst := now.Add(time.Hour * 9)
	nowJst = nowJst.AddDate(0, 0, -1)
	timeDate := fmt.Sprintf("%04d-%02d-%02d 15:00:00", nowJst.Year(), nowJst.Month(), nowJst.Day())

	return timeDate
}

func GetDatePlusDay() string {
	now := time.Now().UTC().AddDate(0, 0, 1)
	timeDate := fmt.Sprintf("%04d-%02d-%02d 00:00:00", now.Year(), now.Month(), now.Day())

	return timeDate
}

func GetTimeDate() string {
	now := time.Now().UTC()
	timeDate := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())

	return timeDate
}
