package gosubmit

import "strconv"

func atoi(str string) int {
	value, _ := strconv.Atoi(str)
	return value
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
