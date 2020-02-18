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
)

type InputOption struct {
	Value    string
	Required bool
}

type Input interface {
	Name() string
	Type() string
	Value() string
	Values() []string
	Options() []InputOption
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
	multiple  bool
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
	opts := i.Options()
	if len(opts) > 0 {
		for _, opt := range opts {
			if opt.Required {
				values = append(values, opt.Value)
			}
		}
		return
	}

	return []string{fmt.Sprintf("%s-%s", i.Name(), "test")}
}

func (i anyInput) Values() []string {
	return i.values
}

func (i anyInput) Required() bool {
	return i.required
}

func (i anyInput) Multiple() bool {
	return i.multiple
}

func (i anyInput) Multipart() bool {
	return false
}

func (i anyInput) Options() (values []InputOption) {
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

type HiddenInput struct {
	anyInput
}

func (i HiddenInput) Fill(val string) (value string, ok bool) {
	return i.Value(), false
}

type inputWithOptions struct {
	anyInput
	options []InputOption
}

func (i inputWithOptions) Options() []InputOption {
	return i.options
}

func (i inputWithOptions) Fill(val string) (value string, ok bool) {
	ok = false
	for _, opt := range i.options {
		if opt.Value == val {
			value = val
			ok = true
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
	return []string{ISO8601Date}
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
