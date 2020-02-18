package gosubmit

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type Option func(f *filler) error

type multipartFile struct {
	Contents []byte
	Name     string
}

type filler struct {
	context     context.Context
	form        Form
	values      url.Values
	url         string
	method      string
	clicked     bool
	multipart   map[string][]multipartFile
	required    map[string]struct{}
	isMultipart bool
}

// Creates a new form filler. It is preferred to use Form.Fill() instead.
func newFiller(form Form, opts []Option) (f *filler, err error) {
	values := make(url.Values)
	f = &filler{
		form:        form,
		values:      values,
		required:    make(map[string]struct{}),
		multipart:   make(map[string][]multipartFile),
		isMultipart: form.ContentType == ContentTypeMultipart,
	}
	f.prefill(form.Inputs)
	err = f.apply(opts)
	return
}

func (f *filler) apply(opts []Option) (err error) {
	for _, opt := range opts {
		err = opt(f)
		if err != nil {
			return
		}
	}
	return
}

func (f *filler) prefill(inputs Inputs) {
	for name, input := range inputs {
		if input.Required() {
			f.required[name] = struct{}{}
		}
		if input.Multipart() {
			continue
		}
		for _, value := range input.Values() {
			f.values.Add(name, value)
		}
	}
}

func (f *filler) createRequest(test bool, method string, url string, body io.Reader) (*http.Request, error) {
	if test {
		return httptest.NewRequest(method, url, body), nil
	}
	ctx := f.context
	if ctx == nil {
		ctx = context.Background()
	}
	return http.NewRequestWithContext(ctx, method, url, body)
}

func (f *filler) NewTestRequest() (*http.Request, error) {
	return f.prepareRequest(true)
}

func (f *filler) NewRequest() (*http.Request, error) {
	return f.prepareRequest(false)
}

// Builds a form depeding on the enctype and creates a new test request.
func (f *filler) prepareRequest(test bool) (r *http.Request, err error) {
	form := f.form
	switch form.Method {
	case http.MethodPost:
		if !f.isMultipart {
			body, err := f.BuildPost()
			if err != nil {
				return nil, err
			}
			r, err = f.createRequest(test, "POST", form.URL, bytes.NewReader(body))
			if err != nil {
				err = fmt.Errorf("Error creating post request: %w", err)
				return nil, err
			}
			r.Header.Add("Content-Type", form.ContentType)
		} else {
			boundary, body, err := f.BuildMultipart()
			if err != nil {
				return nil, err
			}
			r, err = f.createRequest(test, "POST", form.URL, bytes.NewReader(body))
			if err != nil {
				return nil, fmt.Errorf("Error creating multipart request: %w", err)

			}
			r.Header.Add("Content-Type",
				fmt.Sprintf("%s; boundary=%s", ContentTypeMultipart, boundary))
		}
	default:
		query, err := f.BuildGet()
		if err != nil {
			return nil, err
		}
		url := fmt.Sprintf("%s?%s", form.URL, query)
		r, err = f.createRequest(test, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("Error creating get request: %w", err)
		}
	}

	return
}

// Builds form body for a multipart request
func (f *filler) BuildMultipart() (boundary string, data []byte, err error) {
	if err = f.validateForm(); err != nil {
		return "", nil, err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	boundary = writer.Boundary()
	defer func() {
		e := writer.Close()
		if e != nil && err == nil {
			err = fmt.Errorf("Error closing multipart writer: %s", e)
		}
		data = body.Bytes()
	}()

	for field, files := range f.multipart {
		for _, file := range files {
			w, e := writer.CreateFormFile(field, file.Name)
			if e != nil {
				err = fmt.Errorf("Error creating multipart for field '%s': %w", field, e)
				return
			}
			_, err = w.Write(file.Contents)
			if err != nil {
				err = fmt.Errorf("Error writing multipart data for field '%s': %w", field, err)
				return
			}
		}
	}

	for field, values := range f.values {
		for _, value := range values {
			err := writer.WriteField(field, value)
			if err != nil {
				err = fmt.Errorf("Error writing multipart string for field '%s': %w", field, err)
			}
		}
	}

	return
}

func WithContext(ctx context.Context) Option {
	return func(f *filler) error {
		f.context = ctx
		return nil
	}
}

// // Adds value to all empty required fields.
// func (f *filler) AutoFill(defaultValue string) {
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
func (f *filler) validateForm() error {
	for requiredField, _ := range f.required {
		hasTextValue := f.values.Get(requiredField) != ""
		hasByteValue := false
		if f.isMultipart {
			_, hasByteValue = f.multipart[requiredField]
		}
		if !hasTextValue && !hasByteValue {
			return fmt.Errorf("Required field '%s' has no value", requiredField)
		}
	}

	return nil
}

// Build values for form submission
func (f *filler) BuildGet() (params string, err error) {
	err = f.validateForm()
	params = f.values.Encode()
	return params, err
}

// Build form body for post request
func (f *filler) BuildPost() (body []byte, err error) {
	err = f.validateForm()
	body = []byte(f.values.Encode())
	return
}

func AutoFill() Option {
	return func(f *filler) error {
		for requiredField, _ := range f.required {
			value := f.values.Get(requiredField)
			input := f.form.Inputs[requiredField]
			if value == "" {
				add := false
				for _, value := range input.AutoFill() {
					var opt Option
					if input.Type() == InputTypeFile {
						opt = AddFile(requiredField, "auto-filename", []byte(value))
					} else {
						opt = setOrAdd(requiredField, value, add)
					}
					if err := opt(f); err != nil {
						return err
					}
					add = true
				}
			}
		}
		return nil
	}
}

// Adds the submit buttons name=value combination to the form submission.
// Useful when there are two or more buttons on a form and their values
// make a difference on how the server's going to process the form data.
func Click(buttonValue string) Option {
	return func(f *filler) error {
		if f.clicked == true {
			return fmt.Errorf("Already clicked on one button")
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
			return fmt.Errorf("Cannot find button with value: '%s'", buttonValue)
		}
		f.clicked = true
		f.values.Set(b.Name, b.Value)
		return nil
	}
}

// Deletes a field from the form. Useful to remove preselected values
func Reset(name string) Option {
	return func(f *filler) error {
		f.values.Del(name)
		delete(f.multipart, name)
		return nil
	}
}

// Adds a name=value pair to the form. If there is an empty value it will
// be replaced, otherwise a second value will be added, but only if the
// element supports multiple values, like checkboxes or <select multiple>
// elements.
func Add(name string, value string) Option {
	return setOrAdd(name, value, true)
}

// Set a name=value pair to the form and replace any set value(s).
func Set(name string, value string) Option {
	return setOrAdd(name, value, false)
}

func setOrAdd(name string, value string, add bool) Option {
	return func(f *filler) error {
		input, ok := f.form.Inputs[name]
		if !ok {
			return fmt.Errorf("Cannot find input name='%s'", name)
		}
		result, ok := input.Fill(value)
		if !ok {
			return fmt.Errorf("Value '%s' for input name='%s' is invalid", value, name)
		}

		values, ok := f.values[name]
		hasEmptyValue := ok && len(values) == 1 && values[0] == ""

		if add && !hasEmptyValue {
			if f.values.Get(name) != "" && !input.Multiple() {
				return fmt.Errorf("Cannot fill input name='%s'  twice (multiple=false)", name)
			}
			f.values.Add(name, result)
		} else {
			f.values.Set(name, result)
		}
		return nil
	}
}

// Fill data for multipart request
func AddFile(fieldname string, filename string, contents []byte) Option {
	return func(f *filler) error {
		input, ok := f.form.Inputs[fieldname]
		if !ok {
			return fmt.Errorf("Cannot find input fieldname='%s'", fieldname)
		}
		_, ok = input.(FileInput)
		if !ok {
			return fmt.Errorf("Cannot fill bytes - input fieldname='%s' is not a file input", fieldname)
		}
		filesArray, ok := f.multipart[fieldname]
		if !ok {
			filesArray = []multipartFile{}
		}
		f.multipart[fieldname] = append(filesArray, multipartFile{
			Name:     filename,
			Contents: contents,
		})
		return nil
	}
}
