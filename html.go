package gosubmit

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

const ContentTypeForm = "application/x-www-form-urlencoded"
const ContentTypeMultipart = "multipart/form-data"

func Parse(r io.Reader) (Forms, error) {
	n, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("Error parsing html: %w", err)
	}
	forms := findForms(n)
	return forms, nil
}

type Forms map[string]Form
type Inputs map[string]Input

type Form struct {
	// Enctype attribute, default should be application/x-www-form-urlencoded.
	// For forms with file upload it should be multipart/form-data.
	ContentType string
	Inputs      Inputs
	Method      string
	URL         string
	Buttons     []Button
}

var PatternEmail = regexp.MustCompile("[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}$")
var PatternNumber = regexp.MustCompile("^[0-9]+$")

func (f *Form) Submit(name string) {

}

func (f *Form) Fill() *Filler {
	return NewFiller(f.Inputs, f.Buttons)
}

func (f *Form) BuildRequest(filler *Filler) (*http.Request, error) {
	switch f.Method {
	case http.MethodPost:
		switch f.ContentType {
		case ContentTypeForm:
			body, err := filler.Build()
			if err != nil {
				return nil, fmt.Errorf("Error building form request: %w", err)
			}
			r := httptest.NewRequest("POST", f.URL, bytes.NewReader(body))
			r.Header.Add("Content-Type", f.ContentType)
			return r, nil
		case ContentTypeMultipart:
			body, err := filler.BuildMultipart()
			if err != nil {
				return nil, fmt.Errorf("Error building multipart form request: %w", err)
			}
			r := httptest.NewRequest("POST", f.URL, bytes.NewReader(body))
			r.Header.Add("Content-Type", f.ContentType)
			return r, nil
		default:
			return nil, fmt.Errorf("Unknown content type: %s", f.ContentType)
		}
	default:
		query, err := filler.BuildForm()
		if err != nil {
			return nil, err
		}
		return httptest.NewRequest("GET", f.URL+"?"+query, nil), nil
	}
}

func findForms(n *html.Node) (forms Forms) {
	forms = make(Forms)

	var recursivelyFindForms func(n *html.Node)
	recursivelyFindForms = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "form" {
			name := getAttr(n, "name")
			forms[name] = createForm(n)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			recursivelyFindForms(c)
		}
	}
	recursivelyFindForms(n)
	return forms
}

func createAnyInput(n *html.Node) anyInput {
	return anyInput{
		name:      getAttr(n, "name"),
		inputType: getAttr(n, "type"),
		value:     getAttr(n, "value"),
	}
}

func getCheckbox(inputs Inputs, name string) (checkbox Checkbox, ok bool) {
	input, exists := inputs[name]
	ok = exists
	if !ok {
		return
	}
	checkbox, isCheckbox := input.(Checkbox)
	ok = isCheckbox
	return
}

func getRadio(inputs Inputs, name string) (radio Radio, ok bool) {
	input, exists := inputs[name]
	ok = exists
	if !ok {
		return
	}
	radio, isRadio := input.(Radio)
	ok = isRadio
	return
}

func getPattern(n *html.Node) *regexp.Regexp {
	p := getAttr(n, "pattern")
	if p == "" {
		return nil
	}
	return regexp.MustCompile(p)
}

func createForm(n *html.Node) (form Form) {
	inputs := Inputs{}
	var recursivelyFindInputs func(n *html.Node)
	recursivelyFindInputs = func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}
		inputType := getAttr(n, "type")
		name := getAttr(n, "name")
		required := hasAttr(n, "required")
		switch n.Data {
		case "select":
			value, options, _ := findSelectOptions(n)
			inputs[name] = Select{
				inputWithOptions: inputWithOptions{
					anyInput: anyInput{
						name:      name,
						inputType: inputType,
						value:     value,
						required:  required,
						multiple:  hasAttr(n, "multiple"),
					},
					options: options,
				},
			}
		case "input":
			anyInput := anyInput{
				name:      name,
				inputType: inputType,
				value:     getAttr(n, "value"),
				required:  required,
			}
			switch inputType {
			case "checkbox":
				i, ok := getCheckbox(inputs, name)
				if !ok {
					anyInput.multiple = true
					i = Checkbox{
						inputWithOptions: inputWithOptions{
							anyInput: anyInput,
							options:  []string{},
						},
					}
					inputs[name] = i
				}
				// TODO check if this works w/o pointers
				i.options = append(i.options, getAttr(n, "value"))
			case "file":
				inputs[name] = FileInput{
					anyInput: anyInput,
				}
			case "radio":
				i, ok := getRadio(inputs, name)
				if !ok {
					i = Radio{
						inputWithOptions: inputWithOptions{
							anyInput: anyInput,
							options:  []string{},
						},
					}
					inputs[name] = i
				}
				// TODO check if this works w/o pointers
				i.options = append(i.options, getAttr(n, "value"))
			case "submit":
				form.Buttons = append(form.Buttons, Button{
					Name:  name,
					Value: getAttr(n, "value"),
				})
			case "email":
				inputs[name] = TextInput{
					anyInput:  anyInput,
					validator: PatternEmail,
				}
			case "number":
				inputs[name] = TextInput{
					anyInput:  anyInput,
					validator: PatternNumber,
				}
			default:
				inputs[name] = TextInput{
					anyInput:  anyInput,
					validator: getPattern(n),
				}
			}
		case "textarea":
			inputs[name] = TextInput{
				anyInput: anyInput{
					name:      name,
					inputType: "textarea",
					value:     getText(n),
				},
				validator: getPattern(n),
			}
		case "button":
			if inputType == "submit" {
				form.Buttons = append(form.Buttons, Button{
					Name:  name,
					Value: getAttr(n, "value"),
				})
			}
		default:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				recursivelyFindInputs(c)
			}
		}
	}
	recursivelyFindInputs(n)
	form.Inputs = inputs
	form.ContentType = getAttr(n, "enctype")
	if form.ContentType == "" {
		form.ContentType = ContentTypeForm
	}
	form.Method = getAttr(n, "method")
	form.URL = getAttr(n, "action")
	return
}

func getText(n *html.Node) string {
	var b strings.Builder
	var recursivelyGetText func(n *html.Node)
	recursivelyGetText = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			recursivelyGetText(c)
		}
	}
	recursivelyGetText(n)
	return b.String()
}

func findSelectOptions(n *html.Node) (selected string, options []string, ok bool) {
	ok = true
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "option" && !hasAttr(c, "disabled") {
			value := getAttr(c, "value")
			options = append(options, value)
			if hasAttr(c, "selected") {
				selected = value
			}
		}
	}
	return
}

func getAttr(n *html.Node, name string) (value string) {
	value, _ = getAttrOK(n, name)
	return
}

func hasAttr(n *html.Node, name string) bool {
	for _, attr := range n.Attr {
		if attr.Key == name {
			return true
		}
	}
	return false
}

func getAttrOK(n *html.Node, name string) (value string, ok bool) {
	ok = false
	for _, attr := range n.Attr {
		if attr.Key == value {
			ok = true
			value = attr.Val
			break
		}
	}
	return
}
