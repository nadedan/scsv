package tmpl

type Extractor struct {
	Structs []Struct
}

type Struct struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Type string
	Tag  string
}
