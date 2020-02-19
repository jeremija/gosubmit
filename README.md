# gosubmit

[![Build Status](https://travis-ci.com/jeremija/gosubmit.svg?branch=master)](https://travis-ci.com/jeremija/gosubmit)

Helps filling out plain html forms during testing. Will automatically take the
existing values from the form so there is no need to manually set things like
csrf tokens. Alerts about missing required fields, or when pattern validation
does not match. See [example_test.go](example_test.go) for a full example.

```golang
package gosubmit_test

import (
	// TODO import app
	. "github.com/jeremija/gosubmit"

	"net/http"
	"net/http/httptest"
)

func TestLogin(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/login", nil)

	app.ServeHTTP(w, r)

	r, err := ParseResponse(w.Result(), r.URL).FirstForm().NewTestRequest(
		Set("username", "user"),
		Set("password", "password"),
	)

	if err != nil {
		t.Fatalf("Error filling form: %s", err)
	}

	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	if code := w.Result().StatusCode; code != http.StatusOK {
		t.Errorf("Expected status ok but got %d", code)
	}
}
```

Autofilling of all required input fields is supported:

```golang
r, err := ParseResponse(w.Result(), r.URL).FirstForm().NewTestRequest(
	Autofill(),
)
```

Elements that include a pattern attribute for validation will not be autofilled
and have to be filled in manually. For example:

```golang
r, err := ParseResponse(w.Result(), r.URL).FirstForm().NewTestRequest(
	Autofill(),
	Set("validatedURL", "https://www.example.com"),
)
```

# Testing Helpers

To avoid checking for error in tests manually when creating a new test request
, the value of `t *testing.T` can be provided:

```golang
r := ParseResponse(w.Result(), r.URL).FirstForm().Testing(t).NewTestRequest(
	Autofill(),
	Set("validatedURL", "https://www.example.com"),
)
```

In case of any errors, the `t.Fatalf()` function will be called. `t.Helper()`
is used appropriately to ensure line numbers reported by `go test` are correct.

# Supported Elements

- `input[type=checkbox]`
- `input[type=date]`
- `input[type=email]`
- `input[type=hidden]`
- `input[type=number]`
- `input[type=radio]`
- `input[type=text]`
- `input[type=url]`
- `textarea`
- `select`
- `select[multiple]`
- `button[type=submit]` with name and value
- `input[type=submit]` with name and value

If an input element is not on this list, it will default to text input.

# License

MIT
