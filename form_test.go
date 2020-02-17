package gosubmit_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/jeremija/gosubmit"
)

type errReader struct {
	*bytes.Reader
}

func (r *errReader) Read(b []byte) (n int, err error) {
	return 0, fmt.Errorf("A test error")
}

func TestParse_error(t *testing.T) {
	r := &errReader{Reader: bytes.NewReader([]byte("</dflakugk>"))}
	_, err := gosubmit.Parse(r)
	t.Log(err)
	if err == nil {
		t.Error("Expected parsing error, but got nil")
	}
}

func TestParse_Find(t *testing.T) {
	r := bytes.NewReader([]byte("<!DOCTYPE html><html></html>"))
	f, err := gosubmit.Parse(r)
	if err != nil {
		t.Fatalf("Unexpected Parse error: %s", err)
	}
	_, ok := f.Find("name", "test")
	if ok == true {
		t.Fatalf("Expected no forms to be found")
	}
}

func TestParse_GetOptionsFor(t *testing.T) {
	r := bytes.NewReader([]byte(`<!DOCTYPE html>
<html>
<body>
<form>
<input type="checkbox" name="chk" value="one">
<input type="checkbox" name="chk" value="two">
</form>
</html>
`))
	forms, err := gosubmit.Parse(r)
	if err != nil {
		t.Fatalf("Unexpected Parse error: %s", err)
	}
	form := forms[0]
	opts := form.GetOptionsFor("chk")
	if len(opts) != 2 || (opts[0] != "one" && opts[1] != "two") {
		t.Errorf("Expected to find two options")
	}

	opts = form.GetOptionsFor("something-else")
	if len(opts) != 0 {
		t.Errorf("Expected to find no options")
	}
}
