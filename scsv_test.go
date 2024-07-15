package scsv

import (
	"fmt"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	a, err := Parse("")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf(">%s<\n", a.tables[0].name)
	fmt.Printf(">%v<\n", a.tables[0].data[0]["Age"])
}

func TestLocNextTableBanner(t *testing.T) {
	b, err := os.ReadFile("./test/test.scsv")
	if err != nil {
		t.Errorf("could not read file: %s", err)
	}

	s, e := locNextTableBanner(b)

	fmt.Println(string(b[s:e]))
}

func TestAllTableBanners(t *testing.T) {
	b, err := os.ReadFile("./test/test.scsv")
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
	b, err := os.ReadFile("./test/test.scsv")
	if err != nil {
		t.Errorf("could not read file: %s", err)
	}

	s, e := locNextTableBanner(b)

	tableName := parseTableName(b[s:e])

	fmt.Println(tableName)
}
