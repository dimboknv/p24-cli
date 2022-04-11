package p24

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Amount represents p24 amount. It is float value
// with accurate to two decimal places
// for example 123.45, -12, 12.02, 33.20 .
// All number after two decimal places will be skipped during parsing.
// "123.6789" parsed to 123.67
type Amount int64

const (
	// DecimalPrecision of Amount
	DecimalPrecision int64 = 100
)

// String returns the decimal representation of a
func (a Amount) String() string {
	text, _ := a.MarshalText()
	return string(text)
}

// Float64 returns float64 representation of a.
// It returns 0 if a > math.MaxFloat64
func (a Amount) Float64() float64 {
	text, _ := a.MarshalText()
	f, _ := strconv.ParseFloat(string(text), 64)
	return f
}

// MarshalText implements the encoding.TextMarshaler interface for a
func (a Amount) MarshalText() ([]byte, error) {
	str := strconv.FormatInt(int64(a), 10)
	integer, dot, decimal := "", "", ""
	switch {
	case -10 < a && a < 10: //+-[1..9]
		i := len(str) - 1
		integer, dot, decimal = str[:i], "0.0", str[i:]
	case -100 < a && a < 100: //+-[10..99]
		i := len(str) - 2
		integer, dot, decimal = str[:i], "0.", str[i:]
	default:
		i := len(str) - 2
		integer, dot, decimal = str[:i], ".", str[i:]
	}
	str = fmt.Sprintf("%s%s%s", integer, dot, decimal)
	return []byte(strings.Replace(str, ".00", "", 1)), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for a.
// All number after two decimal places will be skipped during parsing.
// "123.6789" parsed to 123.67
func (a *Amount) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return errors.Errorf("parsing %q: invalid syntax", string(text))
	}

	s := string(text) + "00"
	if i := strings.Index(s, "."); i != -1 {
		s = s[:i+3]
		s = strings.Replace(s, ".", "", 1)
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}

	*a = Amount(i)
	return nil
}

// Funds represents p24 funds with special currency code and amount value.
// Funds string representation is "<amount> <currency>", <currency> can be empty string
// for example "23.12 UAH", "-12 USD", "0.0 "
type Funds struct {
	Currency string
	Amount   Amount
}

// MarshalText implements the encoding.TextMarshaler interface for f
func (f Funds) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s %s", f.Amount.String(), f.Currency)), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for f
func (f *Funds) UnmarshalText(text []byte) error {
	vales := bytes.Split(text, []byte(" "))
	if len(vales) != 2 {
		return errors.Errorf("parsing %q: invalid syntax", string(text))
	}

	var amount Amount
	if err := amount.UnmarshalText(vales[0]); err != nil {
		return err
	}
	f.Amount, f.Currency = amount, string(vales[1])

	return nil
}

// String returns string representation of f
func (f Funds) String() string {
	text, _ := f.MarshalText()
	return string(text) // nil slice will be converted to ""
}
