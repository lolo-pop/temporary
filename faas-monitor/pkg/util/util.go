package util

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

func ExtractValueBetween(str, before, after string) (float64, error) {
	envelop := before + `(.*?)` + after
	re := regexp.MustCompile(envelop)
	rs := re.FindStringSubmatch(str)
	stringVal := rs[1]
	val, err := strconv.ParseFloat(stringVal, 64)
	if err != nil {
		msg := fmt.Sprintf("Unable to convert string %v to float\n", stringVal)
		return 0, errors.New(msg)
	}
	return val, nil
}
