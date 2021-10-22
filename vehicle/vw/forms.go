package vw

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/PuerkitoBio/goquery"
)

// FormVars holds HTML form input values required for login
type FormVars struct {
	Action string
	Inputs map[string][]string
}

// FormValues extracts FormVars from given HTML document
func FormValues(reader io.Reader, id string) (FormVars, error) {
	vars := FormVars{Inputs: make(map[string][]string)}

	data, _ := io.ReadAll(reader)
	fmt.Println(string(data))

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err == nil {
		// only interested in meta tag?
		if meta := doc.Find("meta[name=_csrf]"); id == "meta" {
			if meta.Length() != 1 {
				return vars, errors.New("meta not found")
			}

			csrf, exists := meta.Attr("content")
			if !exists {
				return vars, errors.New("meta attribute not found")
			}
			vars.Inputs["_csrf"] = []string{csrf}
			return vars, nil
		}

		form := doc.Find(id).First()
		if form.Length() != 1 {
			return vars, errors.New("form not found")
		}

		action, exists := form.Attr("action")
		if !exists {
			// return vars, errors.New("form attribute not found")
		}
		vars.Action = action

		form.Find("input").Each(func(_ int, el *goquery.Selection) {
			if name, ok := el.Attr("name"); ok {
				if val, ok := el.Attr("value"); ok {
					vars.Inputs[name] = append(vars.Inputs[name], val)
				}
			}
		})
	}

	return vars, err
}
