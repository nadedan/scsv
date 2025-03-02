package scsv

import (
	"fmt"
	"os"
	"testing"
)

func TestUnmarshall(t *testing.T) {
	f, err := os.Open("./testdata/people.scsv")
	if err != nil {
		t.Fatal(err)
	}
	type Thing struct {
		Id     string
		Order  int32
		Weight float32
	}
	type Person struct {
		Name string
		Age  int
	}
	type Data struct {
		Persons []Person
		Things  []Thing
	}

	d := Data{}

	err = Unmarshall(f, &d)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("d: \n%+v", d)
}

func TestParse(t *testing.T) {
	a, err := ParseFile("./testdata/people.scsv")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf(">%s<\n", a.comment)
	fmt.Printf(">%s<\n", a.tables[0].name)
	fmt.Printf(">%v<\n", a.tables[0].rows[0].Value("Age"))

	//	for _, t := range a.Tables() {
	//		t.
	//	}
}

func TestLocNextTableBanner(t *testing.T) {
	b, err := os.ReadFile("./testdata/people.scsv")
	if err != nil {
		t.Errorf("could not read file: %s", err)
	}

	s, e := locNextTableBanner(b)

	fmt.Println(string(b[s:e]))
}

func TestAllTableBanners(t *testing.T) {
	b, err := os.ReadFile("./testdata/people.scsv")
	if err != nil {
		t.Errorf("could not read file: %s", err)
	}

	s, e := locNextTableBanner(b)
	for s < e {
		fmt.Printf("table banner: %s\n", string(b[s:e]))
		b = b[e:]
		//fmt.Println(string(b))
		s, e = locNextTableBanner(b)
		//fmt.Println(string(b[s:e]))
	}

}

func TestTableNameParse(t *testing.T) {
	b, err := os.ReadFile("./testdata/people.scsv")
	if err != nil {
		t.Errorf("could not read file: %s", err)
	}

	s, e := locNextTableBanner(b)

	tableName := parseTableName(b[s:e])

	fmt.Println(tableName)
}
