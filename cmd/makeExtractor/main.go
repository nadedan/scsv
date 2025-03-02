package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path"
	"scsv"
	"strings"
	"text/template"

	"scsv/cmd/makeExtractor/tmpl"
)

func main() {
	fName_p := flag.String("f", "", "Specify the name of the scsv file to process")
	output_p := flag.String("o", "", "Specifty the path of the output package. By default, the output package name comes from the input scsv file")
	flag.Parse()

	var fName string
	switch {
	case len(*fName_p) == 0:
		fName = "../../testdata/people.scsv"
	default:
		fName = *fName_p
	}

	f, err := os.Open(fName)
	if err != nil {
		panic(err)
	}

	var pkgName string
	var pkgDir string
	switch {
	case len(*output_p) == 0:
		p1 := strings.LastIndex(fName, "/") + 1
		p2 := strings.LastIndex(fName, ".")
		pkgName = fName[p1:p2]
		pkgDir = fmt.Sprintf("./%s", pkgName)
	default:
		tmp := strings.TrimRight(*output_p, "/")
		p1 := strings.LastIndex(tmp, "/") + 1
		pkgName = (*output_p)[p1:]
		pkgDir = *output_p
	}

	a, err := scsv.Parse(f)
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

	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(fmt.Sprintf("package %s\n", pkgName))

	err = tmplt.ExecuteTemplate(buf, "extractor.go.tmpl",
		tmpl.Extractor{
			Structs: structs,
		},
	)
	if err != nil {
		panic(err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
		os.MkdirAll(pkgDir, os.ModePerm)
	}

	os.WriteFile(path.Join(pkgDir, pkgName+".go"), formatted, os.ModePerm)

	//fmt.Print(string(formatted))
}
