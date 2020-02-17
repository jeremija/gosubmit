package gosubmit_test

import (
	"io"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/jeremija/gosubmit"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

// func SimpleFormHandler(w http.ResponseWriter, r *http.Request) *http.Request {
// 	r.ParseMultipartForm(defaultMaxMemory)
// 	return r
// }

func mustOpen(t *testing.T, filename string) io.ReadCloser {
	f, err := os.Open("./forms/simple.html")
	if err != nil {
		t.Fatalf("Error opening file: %s (reason: %s)", filename, err)
	}
	return f
}

func TestParseFill_simple(t *testing.T) {
	f := mustOpen(t, "./forms/simple-get.html")
	defer f.Close()

	forms, err := gosubmit.ParseWithURL(f, "/test")
	if err != nil {
		t.Fatal(err)
	}

	form, ok := forms.Find("name", "simple-get")
	if !ok {
		t.Fatal("Could not find form with name=simple-get")
	}

	if form.URL != "/test" {
		t.Fatalf("Expected form url to fallback to /test, but was %s", form.URL)
	}

	r, err := form.Open().
		Add("firstName", "John").
		NewTestRequest()

	if err != nil {
		t.Fatalf("Could not fill form and create test request: %s", err)
	}

	if r.URL.EscapedPath() != "/test" {
		t.Errorf("Expected url path to be /test but was: %s", r.URL.EscapedPath())
	}

	r.ParseForm()
	expected := url.Values{
		"firstName": []string{"John"},
	}
	if !reflect.DeepEqual(expected, r.Form) {
		t.Error("Expected form to be", expected, "but was", r.Form)
	}
}
