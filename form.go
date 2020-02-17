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
	// Enctype attribute, default should be application/x-www-form-urlencoded.
	// For forms with file upload it should be multipart/form-data.
	Attr        []html.Attribute
	ContentType string
	Inputs      Inputs
	Method      string
	URL         string
	Buttons     []Button
}

func (f *Form) Open() *Filler {
	return NewFiller(f)
}
