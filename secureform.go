package secureform

import (
	"mime/multipart"
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
	maxMemory    int64
	maxBytes     int64
	maxStringLen int
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

// NewParser allocates and returns a new Parser with the specified properties.
//
// The memory paramater is the maximum memory used before writing extra to the disk.
// The bytes parameter is the maximum size request body before sending an error
// back to the client. The stringLen parameter is the maximum size string allowed
// in a string form field.
func NewParser(memory, bytes int64, stringLen int) *Parser {
	return &Parser{
		maxMemory:    memory,
		maxBytes:     bytes,
		maxStringLen: stringLen,
	}
}

// Parse parses the form with the options of the parser and loads the results
// into the fields struct.
func (parser *Parser) Parse(w http.ResponseWriter, r *http.Request, fields interface{}) (err error) {
	r.Body = http.MaxBytesReader(w, r.Body, parser.maxBytes)

	err = r.ParseMultipartForm(parser.maxMemory)
	if err != nil {
		return
	}

	err = parser.loadForm(fields, r)
	if err != nil {
		return
	}

	return
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

		err := parser.loadFormValueList(field, &tag, r)
		if err != nil {
			return &FieldError{Name: tag.Name, Err: err}
		}

	}

	return nil
}

func (parser *Parser) loadFormValueList(field reflect.Value, tag *parserTag, r *http.Request) error {
	size := 0
	if isFileField(field) {
		size = len(r.MultipartForm.File[tag.Name])
	} else {
		size = len(r.Form[tag.Name])
	}

	if field.Kind() == reflect.Slice {
		field.Set(reflect.MakeSlice(field.Type(), size, size))
		for i := 0; i < size; i++ {
			err := parser.loadFormValue(field.Index(i), tag, r, i)
			if err != nil {
				return nil
			}
		}
		return nil
	}

	if size == 0 {
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	return parser.loadFormValue(field, tag, r, 0)
}

func (parser *Parser) loadFormValue(field reflect.Value, tag *parserTag, r *http.Request, index int) error {
	switch field.Kind() {
	case reflect.Bool:
		field.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value := formValueByIndex(r.Form, tag.Name, index)
		i, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}

		if err := validateInt(i, tag); err != nil {
			return err
		}

		field.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value := formValueByIndex(r.Form, tag.Name, index)
		u, err := strconv.ParseUint(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}

		if err := validateUint(u, tag); err != nil {
			return err
		}

		field.SetUint(u)

	case reflect.Float32, reflect.Float64:
		value := formValueByIndex(r.Form, tag.Name, index)
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
		value := formValueByIndex(r.Form, tag.Name, index)
		if err := validateString(value, tag, parser.maxStringLen); err != nil {
			return err
		}

		field.SetString(value)

	default:

		// Generic value validator interface.
		if fieldType, ok := field.Addr().Interface().(Type); ok {
			value := formValueByIndex(r.Form, tag.Name, index)
			err := fieldType.Set(value)
			if err != nil {
				return err
			}
			return nil
		}

		// secureform.File struct
		if file, ok := field.Addr().Interface().(*File); ok {
			header := formFileByIndex(r.MultipartForm, tag.Name, index)
			if header == nil {
				return http.ErrMissingFile
			}
			*file = header
			return nil
		}

		return ErrInvalidKind
	}

	return nil
}

func isFileField(field reflect.Value) bool {
	if _, ok := field.Addr().Interface().(*File); ok {
		return true
	}
	if _, ok := field.Addr().Interface().(*[]File); ok {
		return true
	}
	return false
}

func formValueByIndex(form url.Values, name string, index int) string {
	field := form[name]
	if index < len(field) {
		return field[index]
	}
	return ""
}

func formFileByIndex(form *multipart.Form, name string, index int) *multipart.FileHeader {
	field := form.File[name]
	if index < len(field) {
		return field[index]
	}
	return nil
}
