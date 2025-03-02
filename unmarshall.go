package scsv

import (
	"fmt"
	"io"
	"reflect"
)

func Unmarshall(r io.Reader, v any) error {
	a, err := Parse(r)
	if err != nil {
		return fmt.Errorf("scsv.Unmarshall: could not parse r into Archive: %w", err)
	}

	val := reflect.ValueOf(v)
	// Ensure the object is a pointer to a struct
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct")
	}

	// Get the underlying struct's value
	val = val.Elem()
	tableName := val.Type().Name()

	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		fType := val.Type().Field(i)
		if f.Kind() == reflect.Slice {
			unmarshallRows(a, f)
			continue
		}

		table := a.Table(tableName)
		if table == nil {
			panic("nil table")
		}

		rows := table.Rows()
		if len(rows) == 0 {
			panic("no rows")
		}

		f.Set(reflect.ValueOf(rows[0].Value(fType.Name)))
	}

	return nil
}

func unmarshallRows(a Archive, rows reflect.Value) error {
	tableType := rows.Type()
	rowType := tableType.Elem()
	tableName := rowType.Name()

	table := a.Table(tableName)
	if table == nil {
		return fmt.Errorf("unmarshallRows: table %s is not in archive", tableName)
	}

	result := reflect.New(tableType).Elem()
	for _, r := range table.Rows() {
		row := reflect.New(rowType).Elem()
		for i := 0; i < row.NumField(); i++ {
			field := row.Field(i)
			fieldType := row.Type().Field(i)

			if field.Kind() == reflect.Slice {
				err := unmarshallRows(a, field)
				if err != nil {
					return err
				}
				continue
			}

			val := reflect.ValueOf(r.Value(fieldType.Name))
			field.Set(val)
		}

		// Append the element to the result slice
		result = reflect.Append(result, row)
	}

	rows.Set(result)
	return nil
}
