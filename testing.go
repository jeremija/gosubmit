package gosubmit

import (
	"net/http"
)

type test interface {
	Fatalf(format string, values ...interface{})
	Helper()
}

type testingForm struct {
	form Form
	t    test
}

func (f testingForm) assertNoError(err error) {
	f.t.Helper()
	if err != nil {
		f.t.Fatalf("An error occurred: %s", err)
	}
}

func (f testingForm) NewTestRequest(opts ...Option) *http.Request {
	f.t.Helper()
	r, err := f.form.NewTestRequest(opts...)
	f.assertNoError(err)
	return r
}
