package gosubmit

import (
	"fmt"

	"golang.org/x/net/html"
)

type errorContainer struct {
	err error
}

func (e *errorContainer) setError(err error) {
	if e.err == nil {
		e.err = err
	}
}

func (e *errorContainer) Err() error {
	return e.err
}

type Document struct {
	errorContainer
	forms []Form
}

type Inputs map[string]Input

func (d Document) Forms() []Form {
	return d.forms
}

func (d Document) FirstForm() (form Form) {
	if len(d.forms) > 0 {
		return d.forms[0]
	}
	form.URL = "/"
	form.setError(fmt.Errorf("No forms found"))
	return
}

func (d Document) FindForm(attrKey string, attrValue string) (form Form) {
	for _, f := range d.forms {
		for _, attr := range f.Attr {
			if attr.Key == attrKey && attr.Val == attrValue {
				return f
			}
		}
	}

	form.setError(fmt.Errorf("No form with attributes %s='%s' found", attrKey, attrValue))
	return
}

type Form struct {
	errorContainer
	// All html attributes of the form. Used to find the form by attribute
	Attr []html.Attribute
	// Value of Enctype attribute, default is application/x-www-form-urlencoded.
	// For forms with file uploads it should be multipart/form-data.
	ContentType string
	// All found inputs
	Inputs Inputs
	// Value of form method attribute
	Method string
	// Value form action attribute
	URL string
	// All found <button type="submit"> and <input type="submit"> elements.
	Buttons []Button
}

// Creates a new form filler that can be used to fill (and validate) the form
// data before submission.
func (f Form) Fill() *Filler {
	filler := NewFiller(f)
	filler.setError(f.err)
	return filler
}

func (f *Form) GetOptionsFor(name string) (options []string) {
	input, ok := f.Inputs[name]
	if !ok {
		return
	}
	return input.Options()
}
