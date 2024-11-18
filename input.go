package gosubmit

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const (
	AutoFillEmail = "test@example.com"
	AutoFillURL   = "https://www.example.com"
	ISO8601Date   = "2006-01-02"
	AutoFillDate  = ISO8601Date
)

var AutoFillFile = []byte{0xd, 0xe, 0xa, 0xd, 0xb, 0xe, 0xe, 0xf}

type Input interface {
	Name() string
	Type() string
	Value() string
	Values() []string
	Options() []string
	Fill(val string) (value string, ok bool)
	Required() bool
	Multiple() bool
	Multipart() bool
	AutoFill() []string
}

type anyInput struct {
	name      string
	inputType string
	values    []string
	required  bool
}

func (i anyInput) Name() string {
	return i.name
}

func (i anyInput) Type() string {
	return i.inputType
}

func (i anyInput) Value() string {
	if len(i.values) == 0 {
		return ""
	}
	return i.values[0]
}

func (i anyInput) AutoFill() (values []string) {
	return []string{fmt.Sprintf("%s-%s", i.Name(), "autofill")}
}

func (i anyInput) Values() []string {
	return i.values
}

func (i anyInput) Required() bool {
	return i.required
}

func (i anyInput) Multiple() bool {
	return false
}

func (i anyInput) Multipart() bool {
	return false
}

func (i anyInput) Options() (values []string) {
	return
}

type FileInput struct {
	anyInput
}

func (f FileInput) Fill(val string) (value string, ok bool) {
	return "", false
}

func (f FileInput) Multipart() bool {
	return true
}

func (f FileInput) AutoFill() (values []string) {
	values = append(values, string(AutoFillFile))
	return
}

type TextInput struct {
	anyInput
	pattern   *regexp.Regexp
	minLength int
	maxLength int
}

func (i TextInput) Fill(val string) (value string, ok bool) {
	length := len(val)
	ok = i.minLength == 0 && i.maxLength == 0 || length >= i.minLength && length <= i.maxLength
	value = val
	if i.pattern == nil {
		return
	}
	ok = i.pattern.MatchString(value)
	return
}

func (i TextInput) AutoFill() (value []string) {
	length := 10
	if i.pattern != nil {
		return
	}
	if i.minLength > 0 {
		length = i.minLength
		return []string{randomString(length)}
	}
	if i.maxLength > 0 {
		length = i.maxLength
		return []string{randomString(length)}
	}
	return i.anyInput.AutoFill()
}

type HiddenInput struct {
	anyInput
}

func (i HiddenInput) Fill(val string) (value string, ok bool) {
	return i.Value(), false
}

type inputWithOptions struct {
	anyInput
	options  []string
	multiple bool
}

func (i inputWithOptions) Options() []string {
	return i.options
}

func (i inputWithOptions) Multiple() bool {
	return i.multiple
}

func (i inputWithOptions) Fill(val string) (value string, ok bool) {
	ok = false
	for _, opt := range i.options {
		if opt == val {
			value = val
			ok = true
		}
	}
	return
}

func (i inputWithOptions) AutoFill() (values []string) {
	for _, opt := range i.options {
		values = append(values, opt)
		if !i.multiple {
			return
		}
	}
	return
}

type EmailInput struct {
	TextInput
}

func (i EmailInput) AutoFill() []string {
	return []string{AutoFillEmail}
}

type URLInput struct {
	TextInput
}

func (i URLInput) AutoFill() []string {
	return []string{AutoFillURL}
}

type NumberInput struct {
	anyInput
	min int
	max int
}

func (i NumberInput) AutoFill() []string {
	return []string{itoa(i.min)}
}

func (i NumberInput) Fill(val string) (value string, ok bool) {
	ok = false
	intValue, err := strconv.Atoi(val)
	if err != nil {
		return
	}

	ok = i.min == 0 && i.max == 0 || intValue >= i.min && intValue <= i.max
	value = itoa(intValue)
	return
}

type DateInput struct {
	anyInput
}

func (i DateInput) Fill(val string) (value string, ok bool) {
	value = val
	_, err := time.Parse(ISO8601Date, val)
	ok = err == nil
	return
}

func (i DateInput) AutoFill() []string {
	return []string{AutoFillDate}
}

type Checkbox struct {
	inputWithOptions
}

type Radio struct {
	inputWithOptions
}

type Select struct {
	inputWithOptions
}

type Button struct {
	Name  string
	Value string
}
