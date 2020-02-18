package gosubmit_test

import (
	"bytes"
	"fmt"
	"testing"

	. "github.com/jeremija/gosubmit"
)

type errReader struct {
	*bytes.Reader
}

func (r *errReader) Read(b []byte) (n int, err error) {
	return 0, fmt.Errorf("A test error")
}

func TestParse_error(t *testing.T) {
	r := &errReader{Reader: bytes.NewReader([]byte("</dflakugk>"))}
	doc := Parse(r)
	if doc.Err() == nil {
		t.Error("Expected parsing error, but got nil")
	}
}

func TestParse_Find(t *testing.T) {
	r := bytes.NewReader([]byte("<!DOCTYPE html><html></html>"))
	doc := Parse(r)
	if err := doc.Err(); err != nil {
		t.Fatalf("Unexpected Parse error: %s", err)
	}
	form := doc.FindForm("name", "test")
	expected := "No form with attributes name='test' found"
	if err := form.Err(); err == nil || err.Error() != expected {
		t.Fatalf("Expected no error '%s' but got %s", expected, err)
	}
}

func TestParse_GetOptionsFor(t *testing.T) {
	r := bytes.NewReader([]byte(`<!DOCTYPE html>
<html>
<body>
<form>
<input type="checkbox" name="chk" value="one">
<input type="checkbox" name="chk" value="two" required>
</form>
</html>
`))
	doc := Parse(r)
	if err := doc.Err(); err != nil {
		t.Fatalf("Unexpected Parse error: %s", err)
	}
	form := doc.Forms()[0]
	opts := form.GetOptionsFor("chk")
	if len(opts) != 2 || opts[0].Value != "one" || opts[1].Value != "two" {
		t.Errorf("Expected to find two options")
	}

	if len(opts) != 2 || opts[0].Required != false || opts[1].Required != true {
		t.Error("Expected to find two options, first required and second not")
	}

	opts = form.GetOptionsFor("something-else")
	if len(opts) != 0 {
		t.Errorf("Expected to find no options")
	}
}

func TestFirstForm(t *testing.T) {
	var doc Document
	_, err := doc.FirstForm().NewTestRequest(
		Set("a", "b"),
	)

	if err == nil || err.Error() != "No forms found" {
		t.Errorf("Expected an error 'No forms found', but got %s", err)
	}
}
