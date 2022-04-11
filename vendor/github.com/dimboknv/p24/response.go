package p24

import (
	"encoding/xml"
	"io"
	"reflect"

	"github.com/pkg/errors"
)

// ResponseData is struct for mapping p24 response data
type ResponseData struct {
	Info    interface{} `xml:"info"`
	XMLName xml.Name    `xml:"data"`
	Oper    string      `xml:"oper"`
}

// UnmarshalXML implement xml.Unmarshaler interface with decoding to 'Info interface{}' field.
// Info must be not zero value.
// nolint:gocyclo // UnmarshalXML is a complexity operation
func (rd *ResponseData) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if start.Name.Local != "data" {
		return errors.New("invalid start tag")
	}

	var (
		info interface{}
		oper string
	)

	for {
		token, err := d.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		if se, ok := token.(xml.StartElement); ok {
			switch se.Name.Local {
			case "info":
				if rd.Info == nil {
					break
				}
				tmpInfo := reflect.New(reflect.TypeOf(rd.Info)).Interface()
				if err := d.DecodeElement(tmpInfo, &se); err != nil {
					return err
				}
				if !reflect.ValueOf(tmpInfo).Elem().IsZero() {
					info = reflect.ValueOf(tmpInfo).Elem().Interface()
				}
			case "oper":
				if err := d.DecodeElement(&oper, &se); err != nil {
					return err
				}
			}
		}
	}
	if info == nil {
		return errors.New("empty info")
	}
	rd.Info, rd.Oper = info, oper
	return nil
}

// Response is struct for mapping p24 api response
type Response struct {
	Data         ResponseData `xml:"data"`
	XMLName      xml.Name     `xml:"response"`
	MerchantSign MerchantSign `xml:"merchant"`
	Version      string       `xml:"version,attr"`
}

// respDataErr struct for mapping p24 response errors
// nolint:lll // xml example
// like: <?xml version="1.0" encoding="UTF-8"?><response version="1.0"><data><error message ="invalid signature" /></data></response>
type respDataErr struct {
	msg string
}

// Error implements error interface for re
func (re *respDataErr) Error() string {
	return re.msg
}

// UnmarshalXML implement xml.Unmarshaler interface for re
func (re *respDataErr) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	resp := &struct {
		Response
		Data struct {
			Err struct {
				Message string `xml:"message,attr"`
			} `xml:"error"`
		} `xml:"data"`
	}{}

	if err := d.DecodeElement(resp, &start); err != nil {
		return err
	}
	if resp.Data.Err.Message == "" {
		return errors.New("empty error")
	}
	re.msg = resp.Data.Err.Message
	return nil
}

// respDataInfoErr struct for mapping p24 response errors
// <response><data><oper>cmt</oper><info>an error msg</info></data></response>
type respDataInfoErr struct {
	msg string
}

// Error implements error interface for re
func (re *respDataInfoErr) Error() string {
	return re.msg
}

// UnmarshalXML implement xml.Unmarshaler interface for re
// nolint:gocyclo // UnmarshalXML is a complexity operation
func (re *respDataInfoErr) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	resp := &Response{
		Data: ResponseData{
			Info: "",
		},
	}
	if err := d.DecodeElement(resp, &start); err != nil {
		return err
	}
	re.msg = resp.Data.Info.(string)
	return nil
}

// respErr struct for mapping p24 response errors
// like: <error>For input string: "some input"</error>
type respErr struct {
	msg string
}

// Error implements error interface for re
func (re *respErr) Error() string {
	return re.msg
}

// UnmarshalXML implement xml.Unmarshaler interface for re
// nolint:gocyclo // UnmarshalXML is a complexity operation
func (re *respErr) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	resp := &struct {
		XMLName xml.Name `xml:"error"`
		Message string   `xml:",chardata"`
	}{}
	if err := d.DecodeElement(resp, &start); err != nil {
		return err
	}
	if resp.Message == "" {
		return errors.New("empty error")
	}
	re.msg = resp.Message
	return nil
}
