package gosubmit_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
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
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Error opening file: %s (reason: %s)", filename, err)
	}
	return f
}

func TestParseFill_simple_get(t *testing.T) {
	f := mustOpen(t, "./forms/simple.html")
	defer f.Close()

	form := gosubmit.ParseWithURL(f, "/test").FindForm("name", "simple-get")

	if form.URL != "/test" {
		t.Fatalf("Expected form url to fallback to /test, but was %s", form.URL)
	}

	r, err := form.Fill().
		Add("firstName", "John").
		NewTestRequest()

	if err != nil {
		t.Fatalf("Could not fill form and create test request: %s", err)
	}

	if r.Method != http.MethodGet {
		t.Errorf("Expected request method GET but was: %s", r.Method)
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

func TestParseFill_simple_post(t *testing.T) {
	f := mustOpen(t, "./forms/simple.html")
	defer f.Close()

	form := gosubmit.ParseWithURL(f, "/test").FindForm("name", "simple-post")

	if form.URL != "/mytest" {
		t.Fatalf("Expected form url to fallback to /mytest, but was %s", form.URL)
	}

	r, err := form.Fill().
		Add("firstName", "John").
		NewTestRequest()

	if err != nil {
		t.Fatalf("Could not fill form and create test request: %s", err)
	}

	if r.Method != http.MethodPost {
		t.Errorf("Expected request method POST but was: %s", r.Method)
	}

	if r.URL.EscapedPath() != "/mytest" {
		t.Errorf("Expected url path to be /test but was: %s", r.URL.EscapedPath())
	}

	if r.Header.Get("Content-Type") != gosubmit.ContentTypeForm {
		t.Errorf("Expepcted content type to be %s, but was %s",
			gosubmit.ContentTypeForm,
			r.Header.Get("Content-Type"),
		)
	}

	r.ParseForm()
	expected := url.Values{
		"firstName": []string{"John"},
	}
	if !reflect.DeepEqual(expected, r.Form) {
		t.Error("Expected form to be", expected, "but was", r.Form)
	}
}

func TestParseFill_multipart(t *testing.T) {
	f := mustOpen(t, "./forms/big.html")
	defer f.Close()

	form := gosubmit.Parse(f).FirstForm()

	pictureContents := []byte("test-file")

	r, err := form.Fill().
		Set("sel1", form.GetOptionsFor("sel1")[0]).
		Add("sel2", "5").
		// Add("chk", form.GetOptionsFor("chk")[0]).
		Add("chk", form.GetOptionsFor("chk")[1]).
		Set("contact", form.GetOptionsFor("contact")[1]).
		Set("email", "test@example.com").
		Set("firstName", "Test").
		Set("age", "33").
		AddFile("profile", "picture.jpg", pictureContents).
		Click("Save 1").
		NewTestRequest()

	if err != nil {
		t.Fatalf("Error creating test request: %s", err)
	}

	if r.Method != http.MethodPost {
		t.Fatalf("Expected method to be POST, but was %s", r.Method)
	}

	if r.URL.EscapedPath() != "/submit" {
		t.Errorf("Expected url to be /submit but was %s", r.URL.EscapedPath())
	}

	err = r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		t.Fatalf("Error parsing multipart form: %s", err)
	}

	expectedForm := url.Values{
		"sel1":      []string{"1"},
		"sel2":      []string{"4", "6", "5"},
		"chk":       []string{"subscribe-mail", "subscribe-phone"},
		"contact":   []string{"phone"},
		"email":     []string{"test@example.com"},
		"firstName": []string{"Test"},
		"age":       []string{"33"},
		"lastName":  []string{""},
		"csrf":      []string{"1234"},
		"post":      []string{"Big Text"},
		"action":    []string{"Save 1"},
	}

	if !reflect.DeepEqual(expectedForm, r.PostForm) {
		t.Error("Expected form to be:\n", expectedForm, "\nbut was:\n", r.PostForm)
	}

	file, header, err := r.FormFile("profile")
	if err != nil {
		t.Errorf("Cannot read profile image: %s", err)
	}
	defer file.Close()
	if header.Filename != "picture.jpg" {
		t.Errorf("profile filename expected picture.jpg, but was  %s", header.Filename)
	}
	fileData, err := ioutil.ReadAll(file)
	if !bytes.Equal(pictureContents, fileData) {
		t.Errorf("Picture contents do not match: %s vs %s", pictureContents, fileData)
	}
}
