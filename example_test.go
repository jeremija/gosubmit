package gosubmit_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/jeremija/gosubmit"
)

func Serve(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		username := r.FormValue("username")
		password := r.FormValue("password")
		csrf := r.FormValue("csrf")
		fmt.Println(username, password, csrf)
		if csrf == "1234" && username == "user" && password == "pass" {
			w.Write([]byte("Welcome, " + username))
			return
		}
		w.WriteHeader(http.StatusForbidden)
	default:
		w.Write([]byte(`<!DOCTYPE html>
<html>
<body>
<form name="test" method="POST">
	<input type="text" name="username">
	<input type="password" name="password">
	<input type="hidden" name="csrf" value="1234">
	<input type="submit">
</form>`))
	}
}

var mux *http.ServeMux

func init() {
	mux = http.NewServeMux()
	mux.HandleFunc("/auth/login", Serve)
}

func TestLogin(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/login", nil)

	mux.ServeHTTP(w, r)

	forms, _ := gosubmit.ParseResponse(w.Result(), r.URL)
	form := forms[0]

	for _, test := range []struct {
		code int
		pass string
	}{
		// {http.StatusForbidden, "invalid-password"},
		{http.StatusOK, "pass"},
	} {
		t.Run("password_"+test.pass, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, err := form.Fill().
				Set("username", "user").
				Set("password", test.pass).
				NewTestRequest()

			if err != nil {
				t.Fatalf("Error filling in form: %s", err)
			}

			mux.ServeHTTP(w, r)

			if code := w.Result().StatusCode; code != test.code {
				t.Fatalf("Expected status code %d, but got %d", test.code, code)
			}
		})
	}
}

func TestFill_invalid(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/login", nil)

	mux.ServeHTTP(w, r)

	forms, _ := gosubmit.ParseResponse(w.Result(), r.URL)
	form := forms[0]

	_, err := form.Fill().
		Set("invalid-field", "user").
		NewTestRequest()

	re := regexp.MustCompile("Cannot find input name='invalid-field'")
	if err == nil || !re.MatchString(err.Error()) {
		t.Errorf("Expected an error to match %s but got %s", re, err)
	}
}
