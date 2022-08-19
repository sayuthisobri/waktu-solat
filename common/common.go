package common

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

func FindTagValue(t reflect.StructTag, key string) (string, error) {
	if jt, ok := t.Lookup(key); ok {
		return strings.Split(jt, ",")[0], nil
	}
	return "", fmt.Errorf("tag provided does not define a json tag")
}

func ConvertFormatBasedOnTag(t reflect.StructTag, value string) string {
	fromFormat, _ := FindTagValue(t, "fromFormat")
	toFormat, _ := FindTagValue(t, "toFormat")
	if fromFormat != "" && toFormat != "" {
		dateTime, err := time.Parse(fromFormat, value)
		if err == nil {
			return dateTime.Format(toFormat)
		}
	}
	return value
}

func Reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

type Timespan time.Duration

func (ts Timespan) Format() string {
	res := ""
	d := time.Duration(ts)
	calc := func(unit time.Duration, single string, plural string) {
		v := d / unit
		if v > 0 {
			d -= v * unit
			res += fmt.Sprintf("%d%s ", v, Or(v == 1, single, plural))
		}
	}
	calc(time.Hour, "hour", "hours")
	calc(time.Minute, "min", "min")
	calc(time.Second, "sec", "sec")

	return strings.Trim(res, " ")
}

func Or[X any](cond bool, ok X, ko X) X {
	if cond {
		return ok
	}
	return ko
}
