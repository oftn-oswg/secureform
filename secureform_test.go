package secureform

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func NewMockRequest(query string) *http.Request {
	req, err := http.NewRequest("POST", "/", strings.NewReader(query))
	if err != nil {
		panic(err)
	}
	return req
}

type URL url.URL

func (u *URL) Set(value string) error {
	parsed, err := url.Parse(value)
	if err != nil {
		return err
	}
	*u = URL(*parsed)
	return nil
}

func TestParseForm(t *testing.T) {
	tests := []struct {
		Name       string
		Form       interface{}
		Query      string
		ExpErrName string
		ExpErrVal  error
		ExpForm    interface{}
	}{
		{
			Name: "Simple test w/ typo",
			Form: &struct {
				Name string
			}{},
			Query:   "name=foobar",
			ExpForm: &struct{ Name string }{""},
		},
		{
			Name: "Simple test w/ struct tag",
			Form: &struct {
				Name string `form:"name"`
			}{},
			Query: "name=foobar",
			ExpForm: &struct {
				Name string `form:"name"`
			}{"foobar"},
		},
		{
			Name: "Simple test",
			Form: &struct {
				Name string
			}{},
			Query:   "Name=foobar",
			ExpForm: &struct{ Name string }{"foobar"},
		},
		{
			Name:      "Test non-pointer struct",
			Form:      struct{ Name string }{},
			Query:     "Name=foobar",
			ExpErrVal: ErrExpectedStructPtr,
		},
		{
			Name:      "Test pointer int",
			Form:      int(42),
			Query:     "Name=foobar",
			ExpErrVal: ErrExpectedStructPtr,
		},
		{
			Name: "Slice test w/ struct tag",
			Form: &struct {
				Name   string   `form:"name"`
				Fruits []string `form:"fruits"`
			}{},
			Query: "name=foobar;fruits=apple;fruits=banana;fruits=tomato",
			ExpForm: &struct {
				Name   string   `form:"name"`
				Fruits []string `form:"fruits"`
			}{"foobar", []string{"apple", "banana", "tomato"}},
		},
		{
			Name: "Integer min",
			Form: &struct {
				Count int `form:"count?min=1"`
			}{},
			Query:      "count=0",
			ExpErrName: "count",
			ExpErrVal:  ErrValidMin,
		},
		{
			Name: "Tag query param parse error",
			Form: &struct {
				Name string `form:"name?min=1&foo=%zz"`
			}{},
			Query:      "name=Me",
			ExpErrName: "Name",
			ExpErrVal:  nil,
		},
		{
			Name: "Unsigned",
			Form: &struct {
				ID uint `form:"id?max=42"`
			}{},
			Query:      "id=43",
			ExpErrName: "id",
			ExpErrVal:  ErrValidMax,
		},
		{
			Name: "Floating (latitude and longitude)",
			Form: &struct {
				Latitude  float32 `form:"lat?min=-90.0;max=90.0"`
				Longitude float32 `form:"lng?min=-180.0;max=180.0"`
			}{},
			Query: "lat=47.6062;lng=122.3321",
			ExpForm: &struct {
				Latitude  float32 `form:"lat?min=-90.0;max=90.0"`
				Longitude float32 `form:"lng?min=-180.0;max=180.0"`
			}{47.6062, 122.3321},
		},
		{
			Name: "Custom Type (URL)",
			Form: &struct {
				Page URL `form:"page"`
			}{},
			Query:      "page=https://google.com/%25zz",
			ExpErrName: "page",
		},
	}

	for _, test := range tests {
		r := NewMockRequest(test.Query)
		parser := Parser{MaxStringLen: 24}
		query, err := url.ParseQuery(test.Query)
		if err != nil {
			panic(err)
		}
		r.Form = query
		err = parser.loadForm(test.Form, r)
		if err != nil {
			if ferr, ok := err.(*FieldError); ok {
				if ferr.Name != test.ExpErrName {
					t.Errorf("Test %q: Expected error with field %q, got %q", test.Name, test.ExpErrName, ferr.Name)
				}
				if test.ExpErrVal != nil && test.ExpErrVal != ferr.Err {
					t.Errorf("Test %q: Got error %q, expected error %q", test.Name, ferr.Err, test.ExpErrVal)
				}
				continue
			}
			if err != test.ExpErrVal {
				t.Errorf("Test %q: Got error %q, expected error %q", test.Name, err, test.ExpErrVal)
			}
			continue
		}

		if !reflect.DeepEqual(test.Form, test.ExpForm) {
			t.Errorf("Test %q: Expected %#v, got %#v", test.Name, test.ExpForm, test.Form)
			continue
		}

	}
}
