package export

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_DefaultFormatParser(t *testing.T) {
	cases := []struct {
		str      string
		strct    interface{}
		withErr  bool
		expected Format
	}{
		{
			str: "A|B|C",
			strct: struct {
				A int
				B string
				C bool
				a int
			}{},
			expected: Format{
				Str:    "A|B|C",
				Fields: []string{"A", "B", "C"},
				Delim:  ',',
			},
		},
		{
			str: "A|B|C|;",
			strct: struct {
				A int
				B string
				C bool
				a int
			}{},
			expected: Format{
				Str:    "A|B|C|;",
				Fields: []string{"A", "B", "C"},
				Delim:  ';',
			},
		},
		{
			str: "A|B|C",
			strct: struct {
				A int
				B string
				a int
			}{},
			withErr: true,
		},
	}
	for i, c := range cases {
		c := c
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			fp := DefaultFormatParser(c.strct)
			actual, err := MakeFormat(c.str, fp)
			require.True(t, c.withErr == (err != nil), err)
			require.Equal(t, c.expected, actual)
		})
	}
}

func Test_ValuesOf(t *testing.T) {
	type Obj struct {
		A int
		B string
		C bool
	}

	format, err := MakeFormat("A|B|C", DefaultFormatParser(Obj{}))
	require.NoError(t, err)

	cases := []struct {
		obj      interface{}
		expected []interface{}
		withErr  bool
	}{
		{obj: Obj{A: 1, B: "2", C: true}, expected: []interface{}{1, "2", true}},
		{obj: Obj{A: 2, B: "3", C: false}, expected: []interface{}{2, "3", false}},
		{
			obj: struct {
				A  int
				BB string
				C  bool
			}{A: 1, BB: "2", C: true},
			withErr: true,
		},
	}

	for i, c := range cases {
		c := c
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual, err := format.ValuesOf(c.obj)
			require.True(t, c.withErr == (err != nil), err)
			require.Equal(t, c.expected, actual)
		})
	}
}
