package p24

import (
	// nolint:gosec // p24 api require it
	"crypto/md5"
	// nolint:gosec // p24 api require it
	"crypto/sha1"
	"encoding/hex"

	"github.com/pkg/errors"
)

// Merchant store p24 merchant id and password
// see: https://api.privatbank.ua/#p24/registration
type Merchant struct {
	ID   string
	Pass string
}

// MerchantSign store p24 merchant id and signature.
// It required for all p24 responses/requests
type MerchantSign struct {
	ID   string `xml:"id"`
	Sign string `xml:"signature"`
}

// Sign returns MerchantSign of provided data
func (m Merchant) Sign(data []byte) MerchantSign {
	// from documentation:
	// php> $payload=$data.$password
	// php> $sign=sha1(md5($payload))

	payload := []byte(string(data) + m.Pass)
	MD5 := md5.Sum(payload) // nolint:gosec // p24 api require it

	// convert md5 bytes to hex string then to bytes because that how it works in php
	SHA1 := sha1.Sum([]byte(hex.EncodeToString(MD5[:]))) // nolint:gosec // p24 api require it

	return MerchantSign{
		ID:   m.ID,
		Sign: hex.EncodeToString(SHA1[:]),
	}
}

// VerifySign returns an error if dataSign not from m.Sign(data)
func (m Merchant) VerifySign(data []byte, dataSign MerchantSign) error {
	if dataSign != m.Sign(data) {
		return errors.New("invalid signature")
	}
	return nil
}
