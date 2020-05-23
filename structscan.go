package skarner

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrInvalidDest           = errors.New("must pass a pointer to a struct, not a value")
	ErrInvalidEmbeddedStruct = errors.New("embedded struct must have a prefix tag")
)

type rowScanner struct {
	row     *sql.Row
	columns []string
}

func newRowScanner(row *sql.Row, columns []string) *rowScanner {
	return &rowScanner{row: row, columns: columns}
}

func (s *rowScanner) Scan(m interface{}) error {
	model := reflect.ValueOf(m)

	if !isPtr(model) {
		return fmt.Errorf("%w %v", ErrInvalidDest, model.Kind())
	}

	mapValues := make(map[string]interface{})
	sliceValues := make([]interface{}, len(s.columns))

	err := s.row.Scan(sliceValues...)

	for k, column := range s.columns {
		mapValues[column] = sliceValues[k]
	}

	underModel, err := getStructValue(m)
	if err != nil {
		return err
	}

	return structTraversal(underModel, mapValues, "")
}

func isPtr(model reflect.Value) bool {
	return model.Kind() == reflect.Ptr
}

func getStructValue(model interface{}) (reflect.Value, error) {
	modelType := reflect.TypeOf(model)

	if modelType.Kind() != reflect.Ptr {
		return reflect.Value{}, fmt.Errorf("%w %v", ErrInvalidDest, modelType.Kind())
	}

	modelType = modelType.Elem()

	if modelType.Kind() == reflect.Slice {
		modelType = modelType.Elem()
	}

	structValue := reflect.New(modelType)

	return structValue, nil
}

func structTraversal(model reflect.Value, mapValues map[string]interface{}, prefix string) error {
	if err := checkModel(model); err != nil {
		return err
	}

	for i := 0; i < model.Elem().NumField(); i++ {
		fieldValue := reflect.New(model.Elem().Type().Field(i).Type)
		fieldKind := fieldValue.Elem().Kind()
		alias := model.Elem().Type().Field(0).Tag.Get("dbalias")
		column, ok := model.Elem().Type().Field(0).Tag.Lookup("dbcolumn")
		if !ok && !(fieldKind == reflect.Struct) {
			column, ok = model.Elem().Type().Field(0).Tag.Lookup("json")
			if !ok {
				continue
			}
		}

		if fieldKind == reflect.Struct {
			prefixTag, ok := model.Elem().Type().Field(i).Tag.Lookup("prefix")
			if ok {
				err := structTraversal(fieldValue, mapValues, prefixTag)
				if err != nil {
					return err
				}

			} else {
				return ErrInvalidEmbeddedStruct
			}
		}

		columnWithAlias := column
		if prefix == "" && alias != "" {
			columnWithAlias = fmt.Sprintf("%s.%s", alias, column)
		} else if prefix != "" {
			columnWithAlias = fmt.Sprintf("%s.%s", prefix, column)
		}

		value := reflect.Value{}
		if val, ok := mapValues[columnWithAlias]; ok {
			value = reflect.ValueOf(val).Elem()
			if value.IsNil() {
				model.Elem().Field(i).Set(model.Elem().Field(i).Elem())
			}
			model.Elem().Field(i).Set(value)
		}

	}

	return nil
}

func checkModel(model reflect.Value) error {
	if model.Kind() != reflect.Ptr {
		msg := fmt.Sprintf("Expected pointer, got %v", model.Type())
		err := errors.New(msg)
		return err
	}

	model = model.Elem()

	if model.Kind() == reflect.Slice {
		msg := fmt.Sprintf("Expected %v, got %v", model.Type().Elem(), model.Type())
		err := errors.New(msg)
		return err
	}

	return nil
}
