//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"go/format"
	"os"
	"strings"
	"text/template"
)

const categoryTemplate = `// {{ .StructName }} is an auto-generated struct which is used to allow for simple categorisation of the APIs.
// It is public since it may be desired to store a reference to this somewhere, however, do NOT create a instance of this
// directly. Instead, call NewClient and then go to the field {{ .FieldPath }}.
type {{ .StructName }} struct {
	c *Client{{ .AdditionalFields }}
}`

func buildStruct(name, fieldPath, fields string) string {
	if fields != "" {
		fields = "\n\n" + fields
	}
	tpl, err := template.New("tpl").Parse(categoryTemplate)
	if err != nil {
		panic(err)
	}
	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, map[string]string{
		"StructName":       name,
		"FieldPath":        fieldPath,
		"AdditionalFields": fields,
	})
	if err != nil {
		panic(err)
	}
	return buf.String()
}

type field struct {
	key, value, code string
}

func main() {
	// Read categories.json.
	b, err := os.ReadFile("categories.json")
	if err != nil {
		panic(err)
	}
	var categories map[string][]string
	err = json.Unmarshal(b, &categories)
	if err != nil {
		panic(err)
	}

	// Turn the categories into code.
	goCode := []string{}
	for rootCat, subCats := range categories {
		// Get the fields for this category.
		fields := []field{}
		for _, subCat := range subCats {
			structName := subCat
			goName := subCat
			if strings.HasSuffix(goName, ":no-prefix") {
				// Remove this suffix.
				goName = goName[:len(goName)-10]
				structName = goName
			} else {
				// Add the prefix.
				goName = rootCat + goName
			}
			fields = append(fields, field{
				key:   structName,
				value: goName,
				code:  buildStruct(goName, rootCat+"."+structName, ""),
			})
		}

		// Build the root struct.
		fieldsS := ""
		if len(fields) != 0 {
			// Start with 2 blank lines.
			fieldsS += "\n\n"

			// Handle each field.
			for _, v := range fields {
				fieldsS += "\t" + v.key + " *" + v.value + "\n"
			}
		}
		rootStruct := buildStruct(rootCat, rootCat, fieldsS)

		// Append all the structs.
		for _, v := range fields {
			goCode = append(goCode, v.code)
		}
		goCode = append(goCode, rootStruct)

		// Build the root initializer.
		initFunc := "func new" + rootCat + "(c *Client) *" + rootCat + " {"
		if len(fields) == 0 {
			// Do a simple inline init here.
			initFunc += "\n\treturn &" + rootCat + "{c}\n}"
		} else {
			// Handle each field.
			initFunc += "\n\treturn &" + rootCat + "{\n\t\tc: c,"
			for _, v := range fields {
				initFunc += "\n\t\t" + v.key + ": &" + v.value + "{c},"
			}
			initFunc += "\n\t}\n}"
		}
		goCode = append(goCode, initFunc)
	}

	// Make the go file.
	autogen := `// Code generated by generate_categories.go; DO NOT EDIT.
package hopgo

//go:generate go run generate_categories.go`
	for _, code := range goCode {
		autogen += "\n\n" + code
	}
	b, err = format.Source([]byte(autogen))
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("categories_gen.go", b, 0o666)
	if err != nil {
		panic(err)
	}
}