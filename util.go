package gosubmit

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func atoi(str string) int {
	value, _ := strconv.Atoi(str)
	return value
}

func itoa(value int) string {
	return strconv.Itoa(value)
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var randomSource = rand.NewSource(time.Now().UnixNano())

func randomString(size int) string {
	lettersCount := len(letters)
	var b strings.Builder
	b.Grow(size)
	for i := 0; i < size; i++ {
		index := rand.Intn(lettersCount)
		b.WriteByte(letters[index])
	}
	return b.String()
}
