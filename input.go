package gosubmit

import (
	"regexp"
)

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

type TextInput struct {
	anyInput
	validator *regexp.Regexp
}

func (i TextInput) Fill(val string) (value string, ok bool) {
	ok = true
	value = val
	if i.validator == nil {
		return
	}
	ok = i.validator.MatchString(value)
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
	options []string
}

func (i inputWithOptions) Options() []string {
	return i.options
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
