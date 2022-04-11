package export

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/dimboknv/p24"
	"github.com/pkg/errors"
)

var stringer = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

// xmlExporter export statements as xml with custom format
type xmlExporter struct {
	statements p24.Statements
}

// NewXML returns new xmlExporter
func NewXML(statements p24.Statements) Exporter {
	return &xmlExporter{statements}
}

// Export statements to w Writer as xml with given f Format
func (ex *xmlExporter) Export(w io.Writer, f Format) error {
	// encode to temporary buffer for prevent incomplete write
	buff := bytes.NewBuffer([]byte{})
	enc := xml.NewEncoder(buff)
	enc.Indent("", "  ")
	if err := ex.encode(enc, f); err != nil {
		return errors.Wrap(err, "encode failed")
	}
	_, _ = buff.WriteString("\n")

	// write encoded data
	if _, err := w.Write(buff.Bytes()); err != nil {
		return errors.Wrap(err, "failed to write encoded data")
	}
	return nil
}

func (ex *xmlExporter) encode(enc *xml.Encoder, f Format) error {
	// encode top level statements token <statements status="" credit="" debet="">
	if err := enc.EncodeToken(ex.statementsTopLvlStartElem()); err != nil {
		return err
	}

	// encode inner statements list
	// <statement Card="" Appcode="" Trantime="" Trandate="" Amount="" CardAmount="" Rest="" Terminal="" Description=""></statement>
	// ...
	if err := ex.encodeStatementsList(enc, f); err != nil {
		return err
	}

	// close top level statements token
	if err := enc.EncodeToken(ex.statementsTopLvlStartElem().End()); err != nil {
		return err
	}

	return enc.Flush()
}

func (ex *xmlExporter) encodeStatementsList(enc *xml.Encoder, f Format) error {
	stmStartElem := xml.StartElement{
		Name: xml.Name{
			Local: "statement",
		},
		Attr: make([]xml.Attr, len(f.Fields)),
	}
	for i := range f.Fields {
		stmStartElem.Attr[i].Name.Local = strings.ToLower(f.Fields[i])
	}

	updateStmStartElemAttrs := func(sse *xml.StartElement, values []interface{}) {
		for i := 0; i < len(values); i++ {
			v := reflect.ValueOf(values[i])
			if v.CanInterface() && v.Type().Implements(stringer) {
				sse.Attr[i].Value = v.Interface().(fmt.Stringer).String()
				continue
			}
			sse.Attr[i].Value = v.String()
		}
	}
	for i := range ex.statements.Statements {
		values, err := f.ValuesOf(&ex.statements.Statements[i])
		if err != nil {
			return err
		}
		updateStmStartElemAttrs(&stmStartElem, values)
		if err := enc.EncodeToken(stmStartElem); err != nil {
			return err
		}
		if err := enc.EncodeToken(stmStartElem.End()); err != nil {
			return err
		}
	}
	return nil
}

func (ex *xmlExporter) statementsTopLvlStartElem() xml.StartElement {
	return xml.StartElement{
		Name: xml.Name{
			Local: "statements",
		},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "status"}, Value: ex.statements.Status},
			{Name: xml.Name{Local: "credit"}, Value: ex.statements.Credit.String()},
			{Name: xml.Name{Local: "debet"}, Value: ex.statements.Debet.String()},
		},
	}
}
