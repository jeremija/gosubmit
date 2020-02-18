package gosubmit_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"testing"

	. "github.com/jeremija/gosubmit"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

func mustOpen(t *testing.T, filename string) io.ReadCloser {
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Error opening file: %s (reason: %s)", filename, err)
	}
	return f
}

func TestNewTestRequest_simple_get(t *testing.T) {
	f := mustOpen(t, "./forms/simple.html")
	defer f.Close()

	form := ParseWithURL(f, "/test").FindForm("name", "simple-get")

	if form.URL != "/test" {
		t.Fatalf("Expected form url to fallback to /test, but was %s", form.URL)
	}

	r, err := form.NewTestRequest(
		Add("firstName", "John"),
	)

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

func TestFormValidate(t *testing.T) {
	f := mustOpen(t, "./forms/simple.html")
	defer f.Close()

	err := ParseWithURL(f, "/test").FindForm("name", "simple-get").Validate(
		Set("a", "b"),
	)

	expected := "Cannot find input name='a'"
	if err == nil || err.Error() != expected {
		t.Errorf("Expected an error message: '%s', but got '%s'", expected, err.Error())
	}
}

func TestNewTestRequest_simple_post(t *testing.T) {
	f := mustOpen(t, "./forms/simple.html")
	defer f.Close()

	form := ParseWithURL(f, "/test").FindForm("name", "simple-post")

	if form.URL != "/mytest" {
		t.Fatalf("Expected form url to fallback to /mytest, but was %s", form.URL)
	}

	r, err := form.NewTestRequest(
		Add("firstName", "John"),
	)

	if err != nil {
		t.Fatalf("Could not fill form and create test request: %s", err)
	}

	if r.Method != http.MethodPost {
		t.Errorf("Expected request method POST but was: %s", r.Method)
	}

	if r.URL.EscapedPath() != "/mytest" {
		t.Errorf("Expected url path to be /test but was: %s", r.URL.EscapedPath())
	}

	if r.Header.Get("Content-Type") != ContentTypeForm {
		t.Errorf("Expepcted content type to be %s, but was %s",
			ContentTypeForm,
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

func TestNewTestRequest_multipart(t *testing.T) {
	f := mustOpen(t, "./forms/big.html")
	defer f.Close()

	form := Parse(f).FirstForm()

	pictureContents := []byte("test-file")

	r, err := form.NewTestRequest(
		Set("sel1", form.GetOptionsFor("sel1")[0]),
		Add("sel2", "5"),
		// Add("chk", form.GetOptionsFor("chk")[0]),
		Add("chk", form.GetOptionsFor("chk")[1]),
		Set("contact", form.GetOptionsFor("contact")[1]),
		Set("email", "test@example.com"),
		Set("firstName", "Test"),
		Set("age", "33"),
		AddFile("profile", "picture.jpg", pictureContents),
		Click("Save 1"),
	)

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

func TestAutoFill(t *testing.T) {
	f := mustOpen(t, "./forms/big-empty.html")
	defer f.Close()

	form := Parse(f).FirstForm()

	r, err := form.NewTestRequest(
		AutoFill(),
		// autofill won't fill up fields with patterns
		Set("firstName", "John"),
		Click("Save 1"),
	)

	if err != nil {
		t.Fatalf("Error creating test request: %s", err)
	}

	err = r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		t.Fatalf("Error parsing multipart form: %s", err)
	}

	randomLastName := r.FormValue("lastName")

	expectedForm := url.Values{
		"sel2":      []string{"4", "5", "6"},
		"chk":       []string{"subscribe-mail", "subscribe-phone"},
		"contact":   []string{"call"},
		"email":     []string{AutoFillEmail},
		"firstName": []string{"John"},
		"age":       []string{"18"},
		"lastName":  []string{randomLastName},
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
	if header.Filename != "auto-filename" {
		t.Errorf("profile filename expected auto-filename, but was  %s", header.Filename)
	}
	fileData, err := ioutil.ReadAll(file)
	if !bytes.Equal(AutoFillFile, fileData) {
		t.Errorf("Picture contents do not match: %s vs %s", AutoFillFile, fileData)
	}
}

func TestMultipartParams_invalid(t *testing.T) {
	f := mustOpen(t, "./forms/big.html")
	defer f.Close()

	form := Parse(f).FirstForm()

	_, _, err := form.MultipartParams(
		Set("sel1", form.GetOptionsFor("sel1")[0]),
	)

	re := regexp.MustCompile("Required field.*has no value")
	if err == nil || !re.MatchString(err.Error()) {
		t.Errorf("Expected error to match '%s', but was '%s'", re, err.Error())
	}
}

func TestNewRequest_invalid(t *testing.T) {
	f := mustOpen(t, "./forms/simple.html")
	defer f.Close()

	_, err := ParseWithURL(f, "/test").FirstForm().NewRequest(
		Set("a", "b"),
	)

	re := regexp.MustCompile("Cannot find input name='a'")
	if err == nil || !re.MatchString(err.Error()) {
		t.Fatalf("Expected an error %s, but got: %s", re, err)
	}
}

func TestNewRequest_missing(t *testing.T) {
	f := mustOpen(t, "./forms/simple.html")
	defer f.Close()

	_, err := ParseWithURL(f, "/test").FirstForm().NewRequest()

	re := regexp.MustCompile("Required field 'firstName' has no value")
	if err == nil || !re.MatchString(err.Error()) {
		t.Fatalf("Expected an error %s, but got: %s", re, err)
	}
}

func TestFiller_Reset(t *testing.T) {
	f := mustOpen(t, "./forms/simple.html")
	defer f.Close()

	_, err := ParseWithURL(f, "/test").FirstForm().PostParams(
		Set("firstName", "a"),
		Reset("firstName"),
	)

	re := regexp.MustCompile("Required field 'firstName' has no value")
	if err == nil || !re.MatchString(err.Error()) {
		t.Fatalf("Expected an error %s, but got: %s", re, err)
	}
}

func TestFiller_IsRequired(t *testing.T) {
	f := mustOpen(t, "./forms/simple.html")
	defer f.Close()

	isRequired := ParseWithURL(f, "/test").FirstForm().IsRequired("firstName")
	if isRequired == false {
		t.Fatalf("Field 'firstName' is should be required, but is not")
	}
}

func TestFiller_NewRequest(t *testing.T) {
	f := mustOpen(t, "./forms/simple.html")
	defer f.Close()

	ctx := context.WithValue(context.Background(), "a", "b")
	r, err := ParseWithURL(f, "/test").FirstForm().NewRequest(
		Set("firstName", "John"),
		WithContext(ctx),
	)

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if r == nil {
		t.Errorf("Request is nil")
	}

	if c := r.Context(); c != ctx {
		t.Errorf("Context should be the same, but is not")
	}
}

func TestPostParams_error(t *testing.T) {
	f := mustOpen(t, "./forms/simple.html")
	defer f.Close()
	_, err := Parse(f).FirstForm().PostParams(Set("a", "b"))
	expected := "Cannot find input name='a'"
	if err == nil || err.Error() != expected {
		t.Errorf("Expected an error '%s' but got '%s'", expected, err)
	}
}

func TestGetParams_error(t *testing.T) {
	f := mustOpen(t, "./forms/simple.html")
	defer f.Close()
	_, err := Parse(f).FirstForm().GetParams(Set("a", "b"))
	expected := "Cannot find input name='a'"
	if err == nil || err.Error() != expected {
		t.Errorf("Expected an error '%s' but got '%s'", expected, err)
	}
}

func TestGetParams_ok(t *testing.T) {
	f := mustOpen(t, "./forms/simple.html")
	defer f.Close()
	params, err := Parse(f).FirstForm().GetParams(Set("firstName", "b"))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	expected := "firstName=b"
	if params != expected {
		t.Errorf("Expected params to be '%s' but was '%s'", expected, params)
	}
}
