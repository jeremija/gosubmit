package gosubmit

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type multipartFile struct {
	Contents []byte
	Name     string
}

type Filler struct {
	form      *Form
	values    url.Values
	url       string
	method    string
	clicked   bool
	multipart map[string][]multipartFile
	required  map[string]struct{}
	err       error
}

// Creates a new form filler. It is preferred to use Form.Fill() instead.
func NewFiller(form *Form) *Filler {
	values := make(url.Values)
	f := &Filler{
		form:      form,
		values:    values,
		required:  make(map[string]struct{}),
		multipart: make(map[string][]multipartFile),
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
		if input.Multipart() {
			continue
		}
		for _, value := range input.Values() {
			f.values.Add(name, value)
		}
		if input.Required() {
			f.required[name] = struct{}{}
		}
	}
}

// Returns true if field is required, false otherwise.
func (f *Filler) IsFieldRequired(name string) bool {
	_, ok := f.required[name]
	return ok
}

// Builds a form depeding on the enctype and creates a new test request.
func (f *Filler) NewTestRequest() (*http.Request, error) {
	form := f.form
	var r *http.Request
	switch form.Method {
	case http.MethodPost:
		switch form.ContentType {
		case ContentTypeForm:
			body, _ := f.BuildPost()
			r = httptest.NewRequest("POST", form.URL, bytes.NewReader(body))
			r.Header.Add("Content-Type", form.ContentType)
		case ContentTypeMultipart:
			boundary, body, _ := f.BuildMultipart()
			r = httptest.NewRequest("POST", form.URL, bytes.NewReader(body))
			r.Header.Add("Content-Type",
				fmt.Sprintf("%s; boundary=%s", ContentTypeMultipart, boundary))
		default:
			f.setError(fmt.Errorf("Unknown content type: %s", form.ContentType))
		}
	default:
		query, _ := f.BuildGet()
		url := fmt.Sprintf("%s?%s", form.URL, query)
		r = httptest.NewRequest("GET", url, nil)
	}

	return r, f.err
}

// Builds form body for a multipart request
func (f *Filler) BuildMultipart() (boundary string, data []byte, err error) {
	for requiredField, _ := range f.required {
		hasTextValue := f.values.Get(requiredField) != ""
		_, hasByteValue := f.multipart[requiredField]
		if !hasTextValue && !hasByteValue {
			f.setError(fmt.Errorf("Required field '%s' has no value", requiredField))
		}
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	boundary = writer.Boundary()
	defer func() {
		err := writer.Close()
		if err != nil {
			f.setError(fmt.Errorf("Error closing multipart writer: %s", err))
		}
		err = f.err
		data = body.Bytes()
	}()

	for field, files := range f.multipart {
		for _, file := range files {
			w, err := writer.CreateFormFile(field, file.Name)
			if err != nil {
				f.setError(fmt.Errorf("Error creating multipart for field '%s': %w", field, err))
				break
			}
			_, err = w.Write(file.Contents)
			if err != nil {
				f.setError(fmt.Errorf("Error writing multipart data for field '%s': %w", field, err))
			}
		}
	}

	for field, values := range f.values {
		for _, value := range values {
			err := writer.WriteField(field, value)
			if err != nil {
				f.setError(fmt.Errorf("Error writing multipart string for field '%s': %w", field, err))
			}
		}
	}

	return
}

func (f *Filler) Err() error {
	return f.err
}

// // Adds value to all empty required fields.
// func (f *Filler) AutoFill(defaultValue string) {
// 	for requiredField, _ := range f.required {
// 		value := f.values.Get(requiredField)
// 		if value != "" {
// 			continue
// 		}
// 		f.Set(requiredField, fmt.Sprintf("%s-%s", requiredField, defaultValue))
// 	}
// }

// Validates the form (for a plain form request). No need to call this method
// directly if BuildForm or NewTestRequest are used.
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

// Adds the submit buttons name=value combination to the form submission.
// Useful when there are two or more buttons on a form and their values
// make a difference on how the server's going to process the form data.
func (f *Filler) Click(buttonValue string) *Filler {
	if f.clicked == true {
		f.setError(fmt.Errorf("Already clicked on one button"))
		return f
	}
	ok := false
	var b Button
	for _, button := range f.form.Buttons {
		if button.Value == buttonValue {
			ok = true
			b = button
			break
		}
	}
	if !ok {
		f.setError(fmt.Errorf("Cannot find button with value: '%s'", buttonValue))
		return f
	}
	f.clicked = true
	f.values.Set(b.Name, b.Value)
	return f
}

// Deletes a field from the form. Useful to remove preselected values
func (f *Filler) Reset(name string) *Filler {
	f.values.Del(name)
	delete(f.multipart, name)
	return f
}

// Adds a name=value pair to the form. If there is an empty value it will
// be replaced, otherwise a second value will be added, but only if the
// element supports multiple values, like checkboxes or <select multiple>
// elements.
func (f *Filler) Add(name string, value string) *Filler {
	return f.setOrAdd(name, value, true)
}

// Set a name=value pair to the form and replace any set value(s).
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

	values, ok := f.values[name]
	hasEmptyValue := ok && len(values) == 1 && values[0] == ""

	if add && !hasEmptyValue {
		if f.values.Get(name) != "" && !input.Multiple() {
			f.setError(fmt.Errorf("Cannot fill input name='%s'  twice (multiple=false)", name))
			return f
		}
		f.values.Add(name, result)
	} else {
		f.values.Set(name, result)
	}
	return f
}

// Fill data for multipart request
func (f *Filler) AddFile(fieldname string, filename string, contents []byte) *Filler {
	input, ok := f.form.Inputs[fieldname]
	if !ok {
		f.setError(fmt.Errorf("Cannot find input fieldname='%s'", fieldname))
		return f
	}
	_, ok = input.(FileInput)
	if !ok {
		f.setError(fmt.Errorf("Cannot fill bytes - input fieldname='%s' is not a file input", fieldname))
		return f
	}
	filesArray, ok := f.multipart[fieldname]
	if !ok {
		filesArray = []multipartFile{}
	}
	f.multipart[fieldname] = append(filesArray, multipartFile{
		Name:     filename,
		Contents: contents,
	})
	return f
}
