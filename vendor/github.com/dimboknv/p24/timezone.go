package p24

import (
	_ "embed"
	"time"
)

// timeZoneEuropeKiev holds Europe/Kiev the IANA Time Zone database-formatted expectedErrMsg
//go:embed timezone/Kiev
var timeZoneEuropeKiev []byte

var kievLocation = NewKievLocation()

// NewKievLocation returns time.Location of Europe/Kiev time zone
func NewKievLocation() *time.Location {
	l, _ := time.LoadLocationFromTZData("Europe/Kiev", timeZoneEuropeKiev)
	return l
}
