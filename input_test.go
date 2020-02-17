package gosubmit

import (
	"testing"
)

func TestFill(t *testing.T) {
	hi := HiddenInput{}
	_, ok := hi.Fill("test")
	if ok == true {
		t.Errorf("Should not be able to fill in a hidden input")
	}
	fi := FileInput{}
	_, ok = fi.Fill("test")
	if ok == true {
		t.Errorf("Should not be able to fill in a file input")
	}
}

func Test_anyinput(t *testing.T) {
	a := anyInput{
		name:      "test",
		inputType: "checkbox",
		values:    []string{"a", "b"},
	}
	if name := a.Name(); name != a.name {
		t.Errorf("a.Name() should return %s but got %s", a.name, name)
	}
	if inputType := a.Type(); inputType != a.inputType {
		t.Errorf("a.Type() should return %s but got %s", a.inputType, inputType)
	}
	if value := a.Value(); value != a.values[0] {
		t.Errorf("a.Value() should return first value %s but got %s", a.values[0], value)
	}
	if size := len(a.Options()); size > 0 {
		t.Errorf("a.Options() should always return 0 form this type but got %d", size)
	}
}
