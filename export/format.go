package export

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Format represents export format for Exporter.Export func
type Format struct {
	Str    string
	Fields []string
	Delim  rune
}

// String returns Format string representation
func (f Format) String() string {
	return f.Str
}

// FormatParser type is a function that parse str to Format
type FormatParser func(str string) (fields []string, delim rune, err error)

// MakeFormat makes Format. Just set all Format properties received from FormatParser calling
func MakeFormat(str string, fp FormatParser) (Format, error) {
	fields, delim, err := fp(str)
	if err != nil {
		return Format{}, err
	}
	return Format{
		Fields: fields,
		Delim:  delim,
		Str:    str,
	}, nil
}

// ValuesOf returns values of an obj fields in order to f.Fields.
// Returns an error if obj is not struct or pointer to struct
func (f *Format) ValuesOf(obj interface{}) ([]interface{}, error) {
	objVal := reflect.ValueOf(obj)

	switch k := objVal.Kind(); k {
	case reflect.Ptr, reflect.Interface:
		return f.ValuesOf(objVal.Elem().Interface())
	case reflect.Struct:
	default:
		return nil, errors.Errorf("kind %q is not supported. only struct or ptr to struct", k)
	}

	res := make([]interface{}, len(f.Fields))
	for i := 0; i < len(f.Fields); i++ {
		fieldVal := objVal.FieldByName(f.Fields[i])
		if !fieldVal.IsValid() {
			return nil, errors.Errorf("%q field not exist", f.Fields[i])
		}
		if !fieldVal.CanInterface() {
			return nil, errors.Errorf("unexported field %q", f.Fields[i])
		}
		res[i] = fieldVal.Interface()
	}
	return res, nil
}

var nonWordRegexp = regexp.MustCompile(`^\W$`)

// DefaultFormatParser returns FormatParser func where
// input str must be following format: "Field1|Field2|...|FieldN|," or "Field1|Field2|...|FieldN"
// - ',' is delim (default is ',') must be non-word character
// - 'FieldN' name of a valid exported field of struct
func DefaultFormatParser(strct interface{}) FormatParser {
	strctTyp := reflect.TypeOf(strct)
	if k := strctTyp.Kind(); k != reflect.Struct {
		return func(string) ([]string, rune, error) {
			return nil, rune(0), errors.Errorf("kind %q is not supported", k)
		}
	}

	return func(str string) (fields []string, delim rune, err error) {
		// get fields names
		ff := strings.Split(str, "|")
		if len(ff) == 0 { // maybe format str contains one field only
			ff = append(ff, str)
		}

		fields = make([]string, 0, len(ff))
		delim = ',' // default delim

		for i := 0; i < len(ff); i++ {
			// check if latest field string is delim rune
			if i == len(ff)-1 && nonWordRegexp.MatchString(ff[i]) {
				delim = rune(ff[i][0])
				continue
			}

			// check if field name is valid
			if f, ok := strctTyp.FieldByName(ff[i]); !ok || f.Anonymous {
				return nil, rune(0), errors.Errorf("invalid field %q", ff[i])
			}
			fields = append(fields, ff[i])
		}

		if len(fields) == 0 {
			return nil, rune(0), errors.New("no fields")
		}

		return fields, delim, nil
	}
}
