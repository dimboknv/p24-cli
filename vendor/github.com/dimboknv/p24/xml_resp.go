package p24

import (
	"encoding/xml"

	"github.com/pkg/errors"
)

type xmlResp []byte

func (r xmlResp) CheckErr() error {
	respErr := &respErr{}
	if err := xml.Unmarshal(r, respErr); err == nil {
		return respErr
	}

	respDataErr := &respDataErr{}
	if err := xml.Unmarshal(r, respDataErr); err == nil {
		return respDataErr
	}

	respDataInfoErr := &respDataInfoErr{}
	if err := xml.Unmarshal(r, respDataInfoErr); err == nil {
		return respDataInfoErr
	}

	return nil
}

func (r xmlResp) CheckContent() error {
	if _, err := r.commonResp(); err != nil {
		return err
	}

	if _, err := r.dataTagContent(); err != nil {
		return err
	}

	return nil
}

func (r xmlResp) VerifySign(signer Merchant) error {
	dataTag, err := r.dataTagContent()
	if err != nil {
		return err
	}

	resp, err := r.commonResp()
	if err != nil {
		return err
	}

	return signer.VerifySign(dataTag, resp.MerchantSign)
}

func (r xmlResp) dataTagContent() ([]byte, error) {
	dataTag, err := dataTagContent(r)
	if err != nil {
		return nil, errors.Wrap(err, "invalid '<data>' tag")
	}
	return dataTag, nil
}

func (r xmlResp) commonResp() (resp struct {
	Data interface{} `xml:"data"`
	Response
}, err error) {
	if err := xml.Unmarshal(r, &resp); err != nil {
		return resp, errors.Wrap(err, "can`t unmarshal common response")
	}
	return
}
