package scsv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Archive struct {
	comment string
	tables  []*Table
}

type Table struct {
	name    string
	columns []*Column
	rows    []*Row
}

type Column struct {
	name     string
	dataType string
}

type ColumnName = string
type Data = any

type Row map[ColumnName]Data

func (r *Row) set(colName string, value any) {
	(*r)[colName] = value
}

func (r *Row) Value(colName string) any {
	value, ok := (*r)[colName]
	if !ok {
		return nil
	}
	return value
}

const validTypeNames = `
int int8 int16 int32 int64
uint uint8 uint16 uint32 uint64
float32 float64
string`

var fbRe *regexp.Regexp

func init() { fbRe = regexp.MustCompile("\n-- (?P<TableName>.*) --\n") }

func ParseFile(fileName string) (Archive, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return Archive{}, err
	}

	return Parse(f)
}

func Parse(r io.Reader) (Archive, error) {
	a := new(Archive)

	b, err := io.ReadAll(r)
	if err != nil {
		return *a, err
	}

	thisBannerStart, thisBannerEnd := locNextTableBanner(b)
	a.comment = strings.Trim(string(b[:thisBannerStart]), " \n")

	for thisBannerStart < thisBannerEnd {

		t := new(Table)
		t.name = parseTableName(b[thisBannerStart:thisBannerEnd])

		nextBannerStart, _ := locNextTableBanner(b[thisBannerEnd:])

		tableBytes := b[thisBannerEnd : thisBannerEnd+nextBannerStart]

		err = t.load(tableBytes)
		if err != nil {
			return *a, fmt.Errorf("Parse: could not load table %s: %w", t.name, err)
		}

		a.tables = append(a.tables, t)

		b = b[thisBannerEnd:]
		thisBannerStart, thisBannerEnd = locNextTableBanner(b)
	}

	return *a, nil
}

func (a Archive) Tables() []*Table {
	return a.tables
}

func (a Archive) Table(name string) *Table {
	for _, t := range a.tables {
		if name == t.name {
			return t
		}
	}
	return nil
}

func (t *Table) Name() string       { return t.name }
func (t *Table) Columns() []*Column { return t.columns }
func (t *Table) Rows() []*Row       { return t.rows }

func (t *Table) load(b []byte) error {
	rdr := csv.NewReader(bytes.NewReader(b))

	headers, err := rdr.Read()
	if err != nil && err != io.EOF {
		return err
	}
	err = t.loadHeaders(headers)
	if err != nil {
		return fmt.Errorf("%T.loadTable: %w", t, err)
	}

	var row []string
	for err != io.EOF {
		row, err = rdr.Read()
		if err != nil && err != io.EOF {
			return err
		}
		if len(row) == 0 {
			break
		}
		loadRowErr := t.loadRow(row)
		if loadRowErr != nil {
			return fmt.Errorf("%T.loadTable: could not load row %d: %w", t, len(t.rows)+1, loadRowErr)
		}
	}

	return nil
}

func (t *Table) loadHeaders(headers []string) error {

	for _, thisHeader := range headers {
		thisHeader = strings.Trim(thisHeader, " ")

		if strings.ToUpper(string(thisHeader[0])) != string(thisHeader[0]) {
			return fmt.Errorf("%T.loadHeaders: %s is not captalized", t, thisHeader)
		}

		colName, colType, err := decodeHeader(thisHeader)
		if err != nil {
			return fmt.Errorf("%T.loadHeaders: %w", t, err)
		}

		t.columns = append(t.columns, &Column{name: colName, dataType: colType})
	}
	return nil
}

func (t *Table) loadRow(row []string) error {
	if len(row) != len(t.columns) {
		return fmt.Errorf("%T.loadRow: row %v has %d elements but must have %d", t, row, len(row), len(t.columns))
	}

	r := &Row{}
	for i, col := range t.columns {
		dataString := strings.Trim(row[i], " ")

		dataAny, err := col.parse(dataString)
		if err != nil {
			return fmt.Errorf("%T.loadRow: %w", t, err)
		}

		r.set(col.name, dataAny)
	}

	t.rows = append(t.rows, r)

	return nil
}

func (c Column) Name() string { return c.name }
func (c Column) Type() string { return c.dataType }
func (c Column) parse(data string) (any, error) {
	var dataAny any
	var err error
	var i int64
	var u uint64
	var f float64

	switch c.dataType {
	case "int":
		i, err = strconv.ParseInt(data, 0, 64)
		dataAny = int(i)
	case "int8":
		i, err = strconv.ParseInt(data, 0, 8)
		dataAny = int8(i)
	case "int16":
		i, err = strconv.ParseInt(data, 0, 16)
		dataAny = int16(i)
	case "int32":
		i, err = strconv.ParseInt(data, 0, 32)
		dataAny = int32(i)
	case "int64":
		i, err = strconv.ParseInt(data, 0, 64)
		dataAny = int64(i)

	case "uint":
		u, err = strconv.ParseUint(data, 0, 64)
		dataAny = uint(u)
	case "uint8":
		u, err = strconv.ParseUint(data, 0, 8)
		dataAny = uint8(u)
	case "uint16":
		u, err = strconv.ParseUint(data, 0, 16)
		dataAny = uint16(u)
	case "uint32":
		u, err = strconv.ParseUint(data, 0, 32)
		dataAny = uint32(u)
	case "uint64":
		u, err = strconv.ParseUint(data, 0, 64)
		dataAny = uint64(u)

	case "float32":
		f, err = strconv.ParseFloat(data, 32)
		dataAny = float32(f)
	case "float64":
		f, err = strconv.ParseFloat(data, 64)
		dataAny = float64(f)

	case "string":
		dataAny = data
	}

	if err != nil {
		return nil, fmt.Errorf("%T.parseData: %s could not be decoded as type %s", c, data, c.dataType)
	}

	return dataAny, nil
}

func decodeHeader(header string) (colName string, colType string, err error) {
	colName = header
	colType = "string"

	// Regex to match and decode a string that looks like 'ColName(ColType)'
	reColType := regexp.MustCompile(`(?P<ColName>.*)\((?P<ColType>.*)\)`)
	m := reColType.FindStringSubmatch(header)
	if m != nil {
		colName = m[reColType.SubexpIndex("ColName")]
		colType = m[reColType.SubexpIndex("ColType")]
	}

	i := strings.Index(validTypeNames, strings.ReplaceAll(colType, " ", ""))
	if i < 0 {
		return "", "", fmt.Errorf("decodeHeader: %s not a valid column type", colType)
	}

	return colName, colType, nil
}

func locNextTableBanner(b []byte) (start int, end int) {

	loc := fbRe.FindIndex(b)

	if loc == nil {
		return len(b), len(b)
	}
	start = loc[0] + 1
	end = loc[1] - 1

	return start, end
}

func parseTableName(tableBanner []byte) (n string) {

	re := regexp.MustCompile("-- (?P<TableName>.*) --")
	m := re.FindStringSubmatch(string(tableBanner))
	i := re.SubexpIndex("TableName")

	if i > len(m) {
		return ""
	}

	return m[i]
}
