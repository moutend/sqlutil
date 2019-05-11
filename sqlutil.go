package sqlutil

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

func setFields(rows *sql.Rows, m map[string]interface{}) error {
	columns, err := rows.Columns()

	if err != nil {
		return err
	}

	args := make([]interface{}, len(columns))

	for i, column := range columns {
		if v, ok := m[column]; ok {
			args[i] = v
		} else {
			args[i] = struct{}{}
		}
	}
	if err := rows.Scan(args...); err != nil {
		return err
	}

	return nil
}

// mapOf converts struct to map.
func mapOf(i interface{}) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	rv := reflect.ValueOf(i).Elem()

	for n := 0; n < rv.NumField(); n++ {
		if !rv.Field(n).CanSet() {
			continue
		}

		key := rv.Type().Field(n).Name
		m[strings.ToLower(key)] = rv.Field(n).Addr().Interface()
	}

	return m, nil
}

// Bind reads the rows and binds that values to the i.
// Scan accepts struct and slice.
func Bind(rows *sql.Rows, i interface{}) error {
	defer func() {
	}()

	rv := reflect.ValueOf(i).Elem()

	if !rv.CanSet() {
		return fmt.Errorf("brs: specify pointer of i")
	}
	switch rv.Kind() {
	case reflect.Struct:
		return bindStruct(rows, i)
	case reflect.Slice:
		return bindSlice(rows, i)
	default:
		return fmt.Errorf("brs: type %t is not supported", i)
	}

	return nil
}

func bindSlice(rows *sql.Rows, i interface{}) error {
	rv := reflect.ValueOf(i).Elem()

	for rows.Next() {
		cv := reflect.New(rv.Type().Elem())

		m, err := mapOf(cv.Interface())

		if err != nil {
			return err
		}
		if err := setFields(rows, m); err != nil {
			return err
		}

		rv.Set(reflect.Append(rv, cv.Elem()))
	}

	return nil
}

func bindStruct(rows *sql.Rows, i interface{}) error {
	if !rows.Next() {
		return nil
	}
	m, err := mapOf(i)

	if err != nil {
		return err
	}
	if err := setFields(rows, m); err != nil {
		return err
	}

	return nil
}
