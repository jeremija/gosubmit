package gosubmit

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

const ContentTypeForm = "application/x-www-form-urlencoded"
const ContentTypeMultipart = "multipart/form-data"

// Parse all formsr in the HTML document and set the default URL if <form
// action="..."> attribute is missing
func ParseWithURL(r io.Reader, defaultURL string) (doc Document) {
	doc = Parse(r)
	for index, form := range doc.forms {
		if form.URL == "" {
			form.URL = defaultURL
			doc.forms[index] = form
		}
	}
	return
}

// Parse all forms in the HTML document.
func Parse(r io.Reader) (doc Document) {
	n, err := html.Parse(r)
	if err != nil {
		doc.setError(fmt.Errorf("Error parsing html: %w", err))
		return
	}

	doc = findForms(n)
	return
}

func ParseResponse(r *http.Response, url *url.URL) Document {
	return ParseWithURL(r.Body, url.EscapedPath())
}

var PatternEmail = regexp.MustCompile("[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}$")
var PatternURL = regexp.MustCompile("^https?://.+")

func findForms(n *html.Node) (doc Document) {
	var recursivelyFindDocument func(n *html.Node)
	recursivelyFindDocument = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "form" {
			form := createForm(n)
			form.setError(doc.err)
			doc.forms = append(doc.forms, form)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			recursivelyFindDocument(c)
		}
	}
	recursivelyFindDocument(n)
	return doc
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

func getPattern(n *html.Node, defaultPattern *regexp.Regexp) *regexp.Regexp {
	p := getAttr(n, "pattern")
	if p == "" {
		return defaultPattern
	}
	if !strings.HasPrefix(p, "^") {
		p = "^" + p
	}
	if !strings.HasSuffix(p, "$") {
		p = p + "$"
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
			values, options, _ := findSelectOptions(n)
			inputs[name] = Select{
				inputWithOptions: inputWithOptions{
					anyInput: anyInput{
						name:      name,
						inputType: inputType,
						values:    values,
						required:  required,
						multiple:  hasAttr(n, "multiple"),
					},
					options: options,
				},
			}
		case "input":
			value := getAttr(n, "value")
			anyInput := anyInput{
				name:      name,
				inputType: inputType,
				values:    []string{value},
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
							options:  []InputOption{},
						},
					}
					i.values = []string{}
				}
				if hasAttr(n, "checked") {
					i.values = append(i.values, value)
				}
				i.options = append(i.options, InputOption{
					Value:    value,
					Required: hasAttr(n, "required"),
				})
				inputs[name] = i
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
							options:  []InputOption{},
						},
					}
					i.values = []string{}
				}
				i.options = append(i.options, InputOption{
					Value:    value,
					Required: hasAttr(n, "required"),
				})
				if hasAttr(n, "checked") {
					i.values = append(i.values, value)
				}
				// need to reassing because map has plain struct (no pointers)
				inputs[name] = i
			case "hidden":
				inputs[name] = HiddenInput{
					anyInput: anyInput,
				}
			case "submit":
				form.Buttons = append(form.Buttons, Button{
					Name:  name,
					Value: getAttr(n, "value"),
				})
			case "email":
				inputs[name] = EmailInput{
					TextInput: createTextInput(anyInput, n),
				}
			case "url":
				inputs[name] = URLInput{
					TextInput: createTextInput(anyInput, n),
				}
			case "date":
				inputs[name] = DateInput{
					anyInput: anyInput,
				}
			case "number":
				inputs[name] = NumberInput{
					anyInput: anyInput,
					min:      atoi(getAttr(n, "min")),
					max:      atoi(getAttr(n, "max")),
				}
			default:
				inputs[name] = createTextInput(anyInput, n)
			}
		case "textarea":
			inputs[name] = TextInput{
				anyInput: anyInput{
					name:      name,
					inputType: "textarea",
					values:    []string{getText(n)},
				},
				minLength: atoi(getAttr(n, "minlength")),
				maxLength: atoi(getAttr(n, "maxlength")),
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
	form.Method = strings.ToUpper(getAttr(n, "method"))
	if form.Method == "" {
		form.Method = http.MethodGet
	}
	form.URL = getAttr(n, "action")
	form.Attr = n.Attr
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

func findSelectOptions(n *html.Node) (values []string, options []InputOption, ok bool) {
	ok = true
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "option" && !hasAttr(c, "disabled") {
			value := getAttr(c, "value")
			required := hasAttr(c, "required")
			options = append(options, InputOption{Value: value, Required: required})
			if hasAttr(c, "selected") {
				values = append(values, value)
			}
		}
	}
	return
}

func getAttr(n *html.Node, key string) (value string) {
	value, _ = getAttrOK(n, key)
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

func getAttrOK(n *html.Node, key string) (value string, ok bool) {
	ok = false
	for _, attr := range n.Attr {
		if attr.Key == key {
			ok = true
			value = attr.Val
			return
		}
	}
	return
}

func createTextInput(anyInput anyInput, n *html.Node) TextInput {
	return TextInput{
		anyInput:  anyInput,
		pattern:   getPattern(n, nil),
		minLength: atoi(getAttr(n, "minlength")),
		maxLength: atoi(getAttr(n, "maxlength")),
	}
}
