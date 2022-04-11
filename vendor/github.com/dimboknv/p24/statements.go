package p24

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	statementsAPIURL         = "https://api.privatbank.ua/p24api/rest_fiz"
	statementsReqTimeLayout  = "02.01.2006"
	statementsRespDateLayout = "2006-01-02"
	statementsRespTimeLayout = "15:04:05"
)

// StatementsOpts is sets of options required
// for performs p24 statements request
type StatementsOpts struct {
	StartDate  time.Time
	EndDate    time.Time
	CardNumber string
	CommonOpts
}

// Validate r for current date range and card number.
// P24 statements api provide date range with max 90 days
func (r StatementsOpts) Validate() error {
	if r.StartDate.Unix() > r.EndDate.Unix() {
		return errors.New("date range should be with start date <= end date")
	}

	// check date range <= 90 days
	if r.EndDate.Sub(r.StartDate) > 90*24*time.Hour {
		return errors.New("date range should be no longer than 90 days")
	}

	if err := CheckCardNumber(r.CardNumber); err != nil {
		return errors.Wrap(err, "invalid card number")
	}
	return nil
}

// Statements struct for mapping p24 get statements response.
// Represents statements list of a p24 merchant
type Statements struct {
	Status     string      `xml:"status,attr"`
	Statements []Statement `xml:"statement"`
	Credit     Amount      `xml:"credit,attr"`
	Debet      Amount      `xml:"debet,attr"`
}

// Statement represents a Statement of a p24 merchant
type Statement struct {
	Card        string    `xml:"card,attr"`
	Appcode     string    `xml:"appcode,attr"`
	TranDate    time.Time `xml:"-"`
	Terminal    string    `xml:"terminal,attr"`
	Description string    `xml:"description,attr"`
	Amount      Funds     `xml:"amount,attr"`
	CardAmount  Funds     `xml:"cardamount,attr"`
	Rest        Funds     `xml:"rest,attr"`
}

type (
	statementAlias       Statement
	statementXMLMappings struct {
		XMLName     xml.Name `xml:"statement"`
		TranTimeStr string   `xml:"trantime,attr"`
		TranDateStr string   `xml:"trandate,attr"`
		statementAlias
	}
)

// UnmarshalXML implements xml.Unmarshaler interface for s
func (s *Statement) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	statement := &statementXMLMappings{}
	if err := d.DecodeElement(statement, &start); err != nil {
		return err
	}

	// parse tran date/time
	layout := fmt.Sprintf("%s %s", statementsRespDateLayout, statementsRespTimeLayout)
	tranDate, err := time.ParseInLocation(layout, fmt.Sprintf("%s %s", statement.TranDateStr, statement.TranTimeStr), kievLocation)
	if err != nil {
		return err
	}
	*s = Statement(statement.statementAlias)
	s.TranDate = tranDate

	return nil
}

// MarshalXML implements xml.Marshaler interface for s
func (s Statement) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = strings.ToLower(start.Name.Local)
	if start.Name.Local != "statement" {
		return errors.New("invalid start elem name")
	}

	statements := &statementXMLMappings{
		statementAlias: statementAlias(s),
		TranTimeStr:    s.TranDate.Format(statementsRespTimeLayout),
		TranDateStr:    s.TranDate.Format(statementsRespDateLayout),
	}

	return e.EncodeElement(statements, start)
}

// GetStatements returns Statements for given opts.
// Performs p24 orders api call.
// see: https://api.privatbank.ua/#p24/orders
func (c *Client) GetStatements(ctx context.Context, opts StatementsOpts) (Statements, error) {
	if err := opts.Validate(); err != nil {
		return Statements{}, errors.Wrap(err, "invalid request options")
	}

	reqData := RequestData{
		CommonOpts: opts.CommonOpts,
		Payment: struct {
			ID   string "xml:\"id,attr\""
			Prop []struct {
				Name  string "xml:\"name,attr\""
				Value string "xml:\"value,attr\""
			} "xml:\"prop\""
		}{
			Prop: []struct {
				Name  string "xml:\"name,attr\""
				Value string "xml:\"value,attr\""
			}{
				{
					Name:  "sd",
					Value: opts.StartDate.Format(statementsReqTimeLayout),
				},
				{
					Name:  "ed",
					Value: opts.EndDate.Format(statementsReqTimeLayout),
				},
				{
					Name:  "card",
					Value: opts.CardNumber,
				},
			},
		},
	}

	req, err := NewRequest(c.m, reqData)
	if err != nil {
		return Statements{}, errors.Wrap(err, "can`t make request")
	}

	type info struct {
		Statements Statements `xml:"statements"`
	}
	resp := Response{Data: ResponseData{Info: info{}}}
	if err := c.DoContext(ctx, statementsAPIURL, http.MethodPost, req, &resp); err != nil {
		return Statements{}, err
	}

	return resp.Data.Info.(info).Statements, nil
}
