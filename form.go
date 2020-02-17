package gosubmit

import "golang.org/x/net/html"

type Forms []Form
type Inputs map[string]Input

func (ff Forms) Find(attrKey string, attrValue string) (form Form, ok bool) {
	for _, f := range ff {
		for _, attr := range f.Attr {
			if attr.Key == attrKey && attr.Val == attrValue {
				form = f
				ok = true
				return
			}
		}
	}
	return
}

type Form struct {
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
func (f *Form) Fill() *Filler {
	return NewFiller(f)
}

func (f *Form) GetOptionsFor(name string) (options []string) {
	input, ok := f.Inputs[name]
	if !ok {
		return
	}
	return input.Options()
}
