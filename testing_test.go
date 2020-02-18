package gosubmit

import (
	"fmt"
	"regexp"
	"testing"
)

type testMock struct {
	log     []string
	failed  bool
	helpers int
}

func (t *testMock) Fatalf(format string, values ...interface{}) {
	t.log = append(t.log, fmt.Sprintf(format, values...))
	t.failed = true
}

func (t *testMock) Helper() {
	t.helpers++
}

func TestTesting_ok(t *testing.T) {
	var f Form
	f.Inputs = Inputs{}
	input := TextInput{}
	input.name = "firstName"
	input.inputType = "text"
	f.Inputs[input.name] = input
	f.URL = "/test"

	mock := &testMock{}
	f.Testing(mock).NewTestRequest(
		Set("firstName", "John"),
	)

	if mock.failed {
		t.Errorf("Should not fail")
	}

	if len(mock.log) > 0 {
		t.Errorf("Should not write to log")
	}
}

func TestTesting_fail(t *testing.T) {
	var f Form
	f.Inputs = Inputs{}
	input := TextInput{}
	input.name = "firstName"
	input.inputType = "text"
	f.Inputs[input.name] = input
	f.URL = "/test"

	mock := &testMock{}
	f.Testing(mock).NewTestRequest(
		Set("a", "John"),
	)

	if !mock.failed {
		t.Errorf("Should have failed")
	}

	if len(mock.log) != 1 {
		t.Errorf("Should write a log entry")
	}

	if mock.helpers != 2 {
		t.Errorf("Should have marked 2 helpers, but got: %d", mock.helpers)
	}

	log := mock.log[0]
	re := regexp.MustCompile("An error occurred: Cannot find input name='a'")
	if !re.MatchString(log) {
		t.Errorf("Expected log entry to match '%s', but was '%s'", re, log)
	}
}
