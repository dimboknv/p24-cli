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
	cardBalanceAPIURL         = "https://api.privatbank.ua/p24api/balance"
	cardBalanceRespDateLayout = "02.01.06"
	cardBalanceRespTimeLayout = "15:04"
)

// Card represents state of a p24 merchant card
type Card struct {
	Account  string `xml:"account"`
	Number   string `xml:"card_number"`
	AccName  string `xml:"acc_name"`
	AccType  string `xml:"acc_type"`
	Currency string `xml:"currency"`
	Type     string `xml:"card_type"`
	MainCard string `xml:"main_card_number"`
	Status   string `xml:"card_stat"`
	Src      string `xml:"src"`
}

// CardBalance is struct for mapping p24 card balance response.
// Represents balance of a p24 merchant card
type CardBalance struct {
	Date       time.Time `xml:"-"`
	Dyn        string    `xml:"bal_dyn"`
	Card       Card      `xml:"card"`
	Available  Amount    `xml:"av_balance"`
	Balance    Amount    `xml:"balance"`
	FinLimit   Amount    `xml:"fin_limit"`
	TradeLimit Amount    `xml:"trade_limit"`
}

// BalanceOpts is sets of options required
// for performs p24 card balance request
type BalanceOpts struct {
	CardNumber string
	Country    string
	CommonOpts
}

type (
	cardBalanceAlias CardBalance
	cardBalanceXML   struct {
		XMLName xml.Name `xml:"cardbalance"`
		DateStr string   `xml:"bal_date"`
		cardBalanceAlias
	}
)

// UnmarshalXML implements xml.Unmarshaler interface for cb
func (cb *CardBalance) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	cbx := &cardBalanceXML{}
	if err := d.DecodeElement(cbx, &start); err != nil {
		return err
	}

	// parse tran date/time
	layout := fmt.Sprintf("%s %s", cardBalanceRespDateLayout, cardBalanceRespTimeLayout)
	date, err := time.ParseInLocation(layout, cbx.DateStr, kievLocation)
	if err != nil {
		return err
	}
	*cb = CardBalance(cbx.cardBalanceAlias)
	cb.Date = date

	return nil
}

// MarshalXML implements xml.Marshaler interface for cb
func (cb CardBalance) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = strings.ToLower(start.Name.Local)
	if start.Name.Local != "cardbalance" {
		return errors.New("invalid start elem name")
	}

	cbx := &cardBalanceXML{
		cardBalanceAlias: cardBalanceAlias(cb),
		DateStr:          cb.Date.Format(fmt.Sprintf("%s %s", cardBalanceRespDateLayout, cardBalanceRespTimeLayout)),
	}
	return e.EncodeElement(cbx, start)
}

// GetCardBalance returns CardBalance for given opts.
// Performs p24 card balance api call.
// see: https://api.privatbank.ua/#p24/balance
func (c *Client) GetCardBalance(ctx context.Context, opts BalanceOpts) (CardBalance, error) {
	if err := CheckCardNumber(opts.CardNumber); err != nil {
		return CardBalance{}, errors.Wrap(err, "invalid card number")
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
					Name:  "cardnum",
					Value: opts.CardNumber,
				},
				{
					Name:  "country",
					Value: opts.Country,
				},
			},
		},
	}

	req, err := NewRequest(c.m, reqData)
	if err != nil {
		return CardBalance{}, errors.Wrap(err, "can`t make request")
	}

	type info struct {
		CardBalance CardBalance `xml:"cardbalance"`
	}
	resp := Response{Data: ResponseData{Info: info{}}}
	if err := c.DoContext(ctx, cardBalanceAPIURL, http.MethodPost, req, &resp); err != nil {
		return CardBalance{}, err
	}

	return resp.Data.Info.(info).CardBalance, nil
}
