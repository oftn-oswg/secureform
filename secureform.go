package secureform

import (
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Goal: Validate for
// - Type
// - Length
// - Format
// - Range

// Parser defines the security parameters for form parsing.
type Parser struct {
	MaxMemory    int64
	MaxStringLen int
}

type parserTag struct {
	Name     string
	Min, Max string
}

func (t *parserTag) Parse(tag string) error {
	if index := strings.IndexByte(tag, '?'); index >= 0 {
		t.Name = tag[:index]
		values, err := url.ParseQuery(tag[index+1:])
		if err != nil {
			return err
		}
		t.Min = values.Get("min")
		t.Max = values.Get("max")
		return nil
	}
	t.Name = tag
	return nil
}

// Parse parses the form with the options of the parser and loads the results
// into the fields struct.
func (parser *Parser) Parse(fields interface{}, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	if err := parser.loadForm(fields, r); err != nil {
		return err
	}
	return nil
}

// ParseMultipart parses the multipart form with the options of the parser and
// loads the results into the fields struct.
func (parser *Parser) ParseMultipart(fields interface{}, r *http.Request) error {
	if err := r.ParseMultipartForm(parser.MaxMemory); err != nil {
		return err
	}
	if err := parser.loadForm(fields, r); err != nil {
		return err
	}
	return nil
}

func (parser *Parser) loadForm(fields interface{}, r *http.Request) error {
	value := reflect.ValueOf(fields)
	if value.Kind() != reflect.Ptr {
		return ErrExpectedStructPtr
	}
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return ErrExpectedStructPtr
	}

	valueType := value.Type()
	valueTypeLen := valueType.NumField()
	for i := 0; i < valueTypeLen; i++ {
		field := value.Field(i)
		if !field.CanSet() || !field.CanAddr() || !field.CanInterface() {
			continue
		}

		fieldInfo := valueType.Field(i)

		tag := parserTag{Name: fieldInfo.Name}
		if value, ok := fieldInfo.Tag.Lookup("form"); ok {
			err := tag.Parse(value)
			if err != nil {
				return &FieldError{Name: fieldInfo.Name, Err: err}
			}
		}

		err := parser.loadFormValueList(field, &tag, r.Form[tag.Name])
		if err != nil {
			return &FieldError{Name: tag.Name, Err: err}
		}

	}

	return nil
}

func (parser *Parser) loadFormValueList(field reflect.Value, tag *parserTag, list []string) error {
	if field.Kind() == reflect.Slice {
		size := len(list)
		field.Set(reflect.MakeSlice(field.Type(), size, size))
		for i := 0; i < size; i++ {
			err := parser.loadFormValue(field.Index(i), tag, list[i])
			if err != nil {
				return nil
			}
		}
		return nil
	}

	if len(list) == 0 {
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	return parser.loadFormValue(field, tag, list[0])
}

func (parser *Parser) loadFormValue(field reflect.Value, tag *parserTag, value string) error {
	switch field.Kind() {
	case reflect.Bool:
		field.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}

		if err := validateInt(i, tag); err != nil {
			return err
		}

		field.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}

		if err := validateUint(u, tag); err != nil {
			return err
		}

		field.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return err
		}

		if err := validateFloat(f, tag); err != nil {
			return err
		}

		field.SetFloat(f)

	case reflect.String:
		// Validate string
		if err := validateString(value, tag, parser.MaxStringLen); err != nil {
			return err
		}

		field.SetString(value)

	default:

		// Generic value validator interface.
		if fieldType, ok := field.Addr().Interface().(Type); ok {
			err := fieldType.Set(value)
			if err != nil {
				return err
			}
			return nil
		}

		return ErrInvalidKind
	}

	return nil
}
