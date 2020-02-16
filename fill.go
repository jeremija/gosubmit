package gosubmit

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type Filler struct {
	form      *Form
	values    url.Values
	url       string
	method    string
	clicked   bool
	multipart map[string][]byte
	err       error
}

func NewFiller(form *Form) *Filler {
	values := make(url.Values)
	prefill(values, form.Inputs)
	return &Filler{
		form:   form,
		values: values,
	}
}

func prefill(values url.Values, inputs Inputs) {
	for name, input := range inputs {
		for _, value := range input.Values() {
			values.Add(name, value)
		}
	}
}

func (f *Filler) Request(filler *Filler) (*http.Request, error) {
	form := f.form
	switch form.Method {
	case http.MethodPost:
		switch form.ContentType {
		case ContentTypeForm:
			body, err := filler.Build()
			if err != nil {
				return nil, fmt.Errorf("Error building form request: %w", err)
			}
			r := httptest.NewRequest("POST", form.URL, bytes.NewReader(body))
			r.Header.Add("Content-Type", form.ContentType)
			return r, nil
		case ContentTypeMultipart:
			body, err := filler.BuildMultipart()
			if err != nil {
				return nil, fmt.Errorf("Error building multipart form request: %w", err)
			}
			r := httptest.NewRequest("POST", form.URL, bytes.NewReader(body))
			r.Header.Add("Content-Type", form.ContentType)
			return r, nil
		default:
			return nil, fmt.Errorf("Unknown content type: %s", form.ContentType)
		}
	default:
		query, err := filler.BuildForm()
		if err != nil {
			return nil, err
		}
		url := fmt.Sprintf("%s?%s", form.URL, query)
		return httptest.NewRequest("GET", url, nil), nil
	}
}

func (f *Filler) BuildMultipart() ([]byte, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	defer writer.Close()

	for field, data := range f.multipart {
		w, err := writer.CreateFormField(field)
		if err != nil {
			return body.Bytes(), fmt.Errorf("Error creating multipart for field: %s - %w", field, err)
		}
		_, err = w.Write(data)
		if err != nil {
			return body.Bytes(), fmt.Errorf("Error writing multipart data for field: %s - %w", field, err)
		}
	}

	for field, values := range f.values {
		for _, value := range values {
			err := writer.WriteField(field, value)
			if err != nil {
				return body.Bytes(), fmt.Errorf("Error writing multipart string for field: %s - %w", field, err)
			}
		}
	}

	return body.Bytes(), f.err
}

// Build values for form submission
func (f *Filler) BuildForm() (string, error) {
	return f.values.Encode(), f.err
}

// Build form body for post request
func (f *Filler) Build() ([]byte, error) {
	return []byte(f.values.Encode()), f.err
}

func (f *Filler) Click(buttonName string) *Filler {
	if f.clicked == true {
		f.err = fmt.Errorf("Already clicked on one button")
		return f
	}
	ok := false
	var b Button
	for _, button := range f.form.Buttons {
		if button.Name == buttonName {
			ok = true
			b = button
			break
		}
	}
	if !ok {
		f.err = fmt.Errorf("Cannot find button: %s", buttonName)
		return f
	}
	f.clicked = true
	f.values.Set(b.Name, b.Value)
	return f
}

func (f *Filler) Reset(name string) *Filler {
	f.values.Del(name)
	return f
}

func (f *Filler) Fill(name string, value string) *Filler {
	input, ok := f.form.Inputs[name]
	if !ok {
		f.err = fmt.Errorf("Cannot find input name='%s'", name)
		return f
	}
	result, ok := input.Fill(value)
	if !ok {
		f.err = fmt.Errorf("Value '%s' for input name='%s' is invalid", value, name)
		return f
	}
	if f.values.Get(name) != "" && !input.Multiple() {
		f.err = fmt.Errorf("Cannot fill input name='%s'  twice (multiple=false)", name)
		return f
	}
	f.values.Add(name, result)
	return f
}

// Fill data for multipart request
func (f *Filler) FillBytes(name string, bytes []byte) *Filler {
	input, ok := f.form.Inputs[name]
	if !ok {
		f.err = fmt.Errorf("Cannot find input name='%s'", name)
		return f
	}
	_, ok = input.(FileInput)
	if !ok {
		f.err = fmt.Errorf("Cannot fill bytes - input name='%s' is not a file", name)
		return f
	}
	f.multipart[name] = bytes
	return f
}
