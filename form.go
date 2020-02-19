package gosubmit

import (
	"fmt"
	"net/http"

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

// Returns true if field is required, false otherwise.
func (f Form) IsRequired(name string) bool {
	input, ok := f.Inputs[name]
	return ok && input.Required()
}

func (f Form) newFiller(opts []Option) (filler *filler, err error) {
	if f.err != nil {
		err = f.err
		return
	}
	filler, err = newFiller(f, opts)
	return
}

// Fills the form and returns an error if there was an error. Useful for
// testing.
func (f Form) Validate(opts ...Option) error {
	_, err := f.newFiller(opts)
	return err
}

// Fills the form and returns a new request. If there was any error in the
// parsing or if the form was filled incorrectly, it will return an error.
func (f Form) NewRequest(opts ...Option) (*http.Request, error) {
	filler, err := f.newFiller(opts)
	if err != nil {
		return nil, err
	}
	return filler.NewRequest()
}

// Fills the form and returns a new test request. If there was any error in the
// parsing or if the form was filled incorrectly, it will return an error.
func (f Form) NewTestRequest(opts ...Option) (*http.Request, error) {
	filler, err := f.newFiller(opts)
	if err != nil {
		return nil, err
	}
	return filler.NewTestRequest()
}

// Fills the form and returns parameters for a multipart request. If there was
// any error in the parsing or if the form was filled incorrectly, it will
// return an error.
func (f Form) MultipartParams(opts ...Option) (boundary string, data []byte, err error) {
	filler, err := f.newFiller(opts)
	if err != nil {
		return "", nil, err
	}
	return filler.BuildMultipart()
}

// Fills the form and returns query parameters for a GET request.
func (f Form) GetParams(opts ...Option) (string, error) {
	filler, err := f.newFiller(opts)
	if err != nil {
		return "", err
	}
	return filler.BuildGet()
}

// Fills the form and returns body for a POST request.
func (f Form) PostParams(opts ...Option) ([]byte, error) {
	filler, err := f.newFiller(opts)
	if err != nil {
		return nil, err
	}
	return filler.BuildPost()
}

// Returns a list of available input values for elements with options
// (checkbox, radio or select).
func (f Form) GetOptionsFor(name string) (options []string) {
	input, ok := f.Inputs[name]
	if !ok {
		return
	}
	return input.Options()
}

func (f Form) Testing(t test) TestingForm {
	return TestingForm{form: f, t: t}
}
