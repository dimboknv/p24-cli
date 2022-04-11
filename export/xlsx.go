package export

import (
	"fmt"
	"io"

	"github.com/dimboknv/p24"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
)

// xlsxExporter export statements as xlsx with custom format
type xlsxExporter struct {
	xlsx       *excelize.File
	sheet      string
	statements p24.Statements
	row        int
	col        int
	startCol   int
	startRow   int
}

// NewXLSX returns new xlsx exporter
func NewXLSX(statements p24.Statements) Exporter {
	return &xlsxExporter{statements: statements, sheet: "Sheet1", startCol: 2, startRow: 2}
}

// Export statements to w writer as xlsx with given f Format
func (ex *xlsxExporter) Export(w io.Writer, f Format) error {
	if _, err := excelize.CoordinatesToCellName(ex.startCol, ex.startRow); err != nil {
		return err
	}

	// default "Sheet1" created by excelize.NewFile()
	ex.xlsx = excelize.NewFile()
	ex.xlsx.SetActiveSheet(ex.xlsx.NewSheet(ex.sheet))
	ex.row, ex.col = ex.startRow, ex.startCol
	if err := ex.encode(f); err != nil {
		return errors.Wrap(err, "encode failed")
	}

	if err := ex.xlsx.Write(w); err != nil {
		return errors.Wrap(err, "failed to write encoded data")
	}
	return nil
}

func (ex *xlsxExporter) nextRow() {
	ex.row++
	ex.col = ex.startCol
}

func (ex *xlsxExporter) encode(f Format) error {
	// encode Statements table headers
	for i := 0; i < len(f.Fields); i++ {
		if err := ex.encodeHeader(f.Fields[i]); err != nil {
			return err
		}
	}
	ex.nextRow()

	// encode Statements table content
	for i := 0; i < len(ex.statements.Statements); i++ {
		values, err := f.ValuesOf(&ex.statements.Statements[i])
		if err != nil {
			return err
		}
		for k := 0; k < len(values); k++ {
			if err := ex.encodeValue(f.Fields[k], values[k]); err != nil {
				return err
			}
		}
		ex.nextRow()
	}
	return nil
}

func (ex *xlsxExporter) encodeHeader(field string) error {
	switch field {
	case "Amount", "CardAmount", "Rest":
		if err := ex.setCellValue(field); err != nil {
			return err
		}
		if err := ex.setCellValue(fmt.Sprintf("%s Currency", field)); err != nil {
			return err
		}
	default:
		if err := ex.setCellValue(field); err != nil {
			return err
		}
	}
	return nil
}

func (ex *xlsxExporter) setCellValue(value interface{}) error {
	if err := ex.xlsx.SetCellValue(ex.sheet, ex.axis(), value); err != nil {
		return err
	}
	ex.col++
	return nil
}

func (ex *xlsxExporter) encodeValue(field string, value interface{}) error {
	switch field {
	case "Amount", "CardAmount", "Rest":
		amount := value.(p24.Funds)
		if err := ex.setCellValue(amount.Amount.Float64()); err != nil {
			return err
		}

		if err := ex.setCellValue(amount.Currency); err != nil {
			return err
		}
	default:
		if err := ex.setCellValue(value); err != nil {
			return err
		}
	}
	return nil
}

func (ex *xlsxExporter) axis() string {
	str, _ := excelize.CoordinatesToCellName(ex.col, ex.row)
	return str
}
