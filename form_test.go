package gosubmit_test

import (
	"bytes"
	"fmt"
	"reflect"
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
func TestFindFormsByClass(t *testing.T) {
	r := bytes.NewReader([]byte(`<!DOCTYPE html>
<html>
<body>
<form class="a b">
<input type="checkbox" name="one-chk" value="two" required>
</form>
<form class="b c">
<input type="checkbox" name="two-chk" value="one">
</form>
</html>
`))
	doc := Parse(r)
	if err := doc.Err(); err != nil {
		t.Fatalf("Unexpected Parse error: %s", err)
	}
	forms := doc.FindFormsByClass("a")
	if size := len(forms); size != 1 {
		t.Fatalf("Expected to find one form with class a, but got: %d", size)
	}
	if !reflect.DeepEqual(forms.First(), forms.Last()) {
		t.Fatalf("Expected first and last to be same")
	}
	forms = doc.FindFormsByClass("b")
	if size := len(forms); size != 2 {
		t.Fatalf("Expected to find two forms with class b, but got: %d", size)
	}
	if reflect.DeepEqual(forms.First(), forms.Last()) {
		t.Fatalf("Expected forms to be different")
	}
	forms = doc.FindFormsByClass("c")
	if size := len(forms); size != 1 {
		t.Fatalf("Expected to find two forms with class c, but got: %d", size)
	}
	if !reflect.DeepEqual(forms.First(), forms.Last()) {
		t.Fatalf("Expected first and last to be same")
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
	if len(opts) != 2 || opts[0] != "one" || opts[1] != "two" {
		t.Errorf("Expected to find two options")
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
