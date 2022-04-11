package p24

import (
	"bytes"
	"regexp"

	"github.com/pkg/errors"
)

var onlyNumbers = regexp.MustCompile(`^\d+$`)

// CheckCardNumber returns an error if card number is not valid
func CheckCardNumber(card string) error {
	switch {
	case len(card) != 16:
		return errors.New("should be sixteen length")
	case !onlyNumbers.MatchString(card):
		return errors.New("should contains digits only")
	default:
		return nil
	}
}

func dataTagContent(data []byte) ([]byte, error) {
	start, end := bytes.Index(data, []byte("<data>")), bytes.LastIndex(data, []byte("</data>"))
	if start == -1 || end == -1 {
		return nil, errors.New("not found")
	}
	start += len("<data>")
	cnt := make([]byte, end-start)
	copy(cnt, data[start:end])
	return cnt, nil
}
