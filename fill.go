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
	required  map[string]struct{}
	err       error
}

func NewFiller(form *Form) *Filler {
	values := make(url.Values)
	f := &Filler{
		form:     form,
		values:   values,
		required: make(map[string]struct{}),
	}
	f.prefill(form.Inputs)
	return f
}

func (f *Filler) setError(err error) {
	if f.err == nil {
		f.err = err
	}
}

func (f *Filler) prefill(inputs Inputs) {
	for name, input := range inputs {
		for _, value := range input.Values() {
			f.values.Add(name, value)
		}
		if input.Required() {
			f.required[name] = struct{}{}
		}
	}
}

func (f *Filler) IsFieldRequired(name string) bool {
	_, ok := f.required[name]
	return ok
}

func (f *Filler) NewTestRequest() (*http.Request, error) {
	form := f.form
	var r *http.Request
	switch form.Method {
	case http.MethodPost:
		switch form.ContentType {
		case ContentTypeForm:
			body, _ := f.BuildPost()
			r = httptest.NewRequest("POST", form.URL, bytes.NewReader(body))
		case ContentTypeMultipart:
			body, _ := f.BuildMultipart()
			r = httptest.NewRequest("POST", form.URL, bytes.NewReader(body))
		default:
			f.setError(fmt.Errorf("Unknown content type: %s", form.ContentType))
		}
	default:
		query, _ := f.BuildGet()
		url := fmt.Sprintf("%s?%s", form.URL, query)
		r = httptest.NewRequest("GET", url, nil)
	}

	if r != nil {
		r.Header.Add("Content-Type", form.ContentType)
	}

	return r, f.err
}

func (f *Filler) BuildMultipart() ([]byte, error) {
	for requiredField, _ := range f.required {
		hasTextValue := f.values.Get(requiredField) != ""
		_, hasByteValue := f.multipart[requiredField]
		if !hasTextValue && !hasByteValue {
			f.setError(fmt.Errorf("Required field '%s' has no value", requiredField))
		}
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	defer writer.Close()

	for field, data := range f.multipart {
		w, err := writer.CreateFormField(field)
		if err != nil {
			return body.Bytes(), fmt.Errorf("Error creating multipart for field '%s': %w", field, err)
		}
		_, err = w.Write(data)
		if err != nil {
			return body.Bytes(), fmt.Errorf("Error writing multipart data for field '%s': %w", field, err)
		}
	}

	for field, values := range f.values {
		for _, value := range values {
			err := writer.WriteField(field, value)
			if err != nil {
				return body.Bytes(), fmt.Errorf("Error writing multipart string for field '%s': %w", field, err)
			}
		}
	}

	return body.Bytes(), f.err
}

// Validates the form (for a plain form request)
func (f *Filler) ValidateForm() error {
	for requiredField, _ := range f.required {
		value := f.values.Get(requiredField)
		if value == "" {
			f.setError(fmt.Errorf("ValidateForm: required field '%s' has no value", requiredField))
		}
	}
	return f.err
}

// Build values for form submission
func (f *Filler) BuildGet() (string, error) {
	f.ValidateForm()
	return f.values.Encode(), f.err
}

// Build form body for post request
func (f *Filler) BuildPost() ([]byte, error) {
	f.ValidateForm()
	return []byte(f.values.Encode()), f.err
}

func (f *Filler) Click(buttonName string) *Filler {
	if f.clicked == true {
		f.setError(fmt.Errorf("Already clicked on one button"))
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
		f.setError(fmt.Errorf("Cannot find button: %s", buttonName))
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

func (f *Filler) Add(name string, value string) *Filler {
	return f.setOrAdd(name, value, true)
}

func (f *Filler) Set(name string, value string) *Filler {
	return f.setOrAdd(name, value, false)
}

func (f *Filler) setOrAdd(name string, value string, add bool) *Filler {
	input, ok := f.form.Inputs[name]
	if !ok {
		f.setError(fmt.Errorf("Cannot find input name='%s'", name))
		return f
	}
	result, ok := input.Fill(value)
	if !ok {
		f.setError(fmt.Errorf("Value '%s' for input name='%s' is invalid", value, name))
		return f
	}
	if f.values.Get(name) != "" && !input.Multiple() {
		f.setError(fmt.Errorf("Cannot fill input name='%s'  twice (multiple=false)", name))
		return f
	}

	values, ok := f.values[name]
	hasEmptyValue := ok && len(values) == 1 && values[0] == ""

	if add && !hasEmptyValue {
		f.values.Add(name, result)
	} else {
		f.values.Set(name, result)
	}
	logger.Printf("Added value for field '%s': '%s'", name, result)
	return f
}

// Fill data for multipart request
func (f *Filler) FillBytes(name string, bytes []byte) *Filler {
	input, ok := f.form.Inputs[name]
	if !ok {
		f.setError(fmt.Errorf("Cannot find input name='%s'", name))
		return f
	}
	_, ok = input.(FileInput)
	if !ok {
		f.setError(fmt.Errorf("Cannot fill bytes - input name='%s' is not a file", name))
		return f
	}
	f.multipart[name] = bytes
	return f
}
