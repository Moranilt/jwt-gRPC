package utils

import (
	"fmt"
	"strconv"
	"time"
)

const (
	ERROR_NotValidUnit        = "not valid unit %q. Expected d, h, m, s"
	ERROR_NotValidNumberValue = "not valid number value %q"
	ERROR_NotValidQuantity    = "not valid quantity of time. Got %q, Expected number greater than 0"
	ERROR_NotValidTime        = "not valid time value"
)

func MakeTimeFromString(st string) (time.Duration, error) {
	if len(st) < 2 {
		return time.Nanosecond, fmt.Errorf(ERROR_NotValidTime)
	}

	start, end := st[:len(st)-1], st[len(st)-1:]

	startInt, err := strconv.Atoi(start)
	if err != nil {
		return time.Nanosecond, fmt.Errorf(ERROR_NotValidNumberValue, start)
	}

	if startInt == 0 {
		return time.Nanosecond, fmt.Errorf(ERROR_NotValidQuantity, start)
	}

	switch end {
	case "d":
		return time.Duration(startInt) * 24 * time.Hour, nil
	case "h":
		return time.Duration(startInt) * time.Hour, nil
	case "m":
		return time.Duration(startInt) * time.Minute, nil
	case "s":
		return time.Duration(startInt) * time.Second, nil
	default:
		return time.Nanosecond, fmt.Errorf(ERROR_NotValidUnit, end)
	}
}
