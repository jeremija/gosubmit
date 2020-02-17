# gosubmit

[![Build Status](https://travis-ci.com/jeremija/gosubmit.svg?branch=master)](https://travis-ci.com/jeremija/gosubmit)

Helps filling out plain html during testing. Will automatically take the
existing values from the form so there is no need to manually set things like
csrf tokens. Alerts about missing required fields, or when pattern validation
does not match. See [example_test.go](example_test.go) for a full example.

```golang
package gosubmit_test

import (
	// TODO import app
	"github.com/jeremija/gosubmit"

	"net/http"
	"net/http/httptest"
)

func TestLogin(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/login", nil)

	app.ServeHTTP(w, r)

	forms, _ := gosubmit.ParseResponse(w.Result(), r.URL)
	r, err := forms[0].Fill().
		Set("username", "user").
		Set("password", "password").
		NewTestRequest()

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

Currently supported elements:

- input[type=text]
- input[type=number]
- input[type=email]
- input[type=checkbox]
- input[type=radio]
- input[type=hidden]
- textarea
- select
- select[multiple]
- button[type=submit] with name and value
- input[type=submit] with name and value

If an input element is not on this list, it will default to text input.

# License

MIT
