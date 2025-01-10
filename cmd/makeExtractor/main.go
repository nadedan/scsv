package main

import (
	"os"
	"scsv"
	"text/template"

	"scsv/cmd/makeExtractor/tmpl"
)

func main() {
	a, err := scsv.ParseFile("../../test/test.scsv")
	if err != nil {
		panic(err)
	}

	var _ template.Template
	tmplt := template.Must(template.ParseGlob("./tmpl/*.go.tmpl"))

	structs := make([]tmpl.Struct, 0)
	for _, t := range a.Tables() {

		fields := make([]tmpl.Field, 0)
		for _, c := range t.Columns() {
			fields = append(fields, tmpl.Field{
				Name: c.Name(),
				Type: c.Type(),
			})
		}

		structs = append(structs, tmpl.Struct{
			Name:   t.Name(),
			Fields: fields,
		})

	}
	err = tmplt.ExecuteTemplate(os.Stdout, "extractor.go.tmpl",
		tmpl.Extractor{
			Structs: structs,
		},
	)
	if err != nil {
		panic(err)
	}

}
