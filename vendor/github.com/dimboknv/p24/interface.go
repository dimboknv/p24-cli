package p24

import "net/http"

// Doer defines minimal http client interface
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// DoFunc type is an adapter to allow the use of ordinary functions as Doer
type DoFunc func(req *http.Request) (*http.Response, error)

// Do calls do(req)
func (do DoFunc) Do(req *http.Request) (*http.Response, error) { return do(req) }

// Logger defines minimal logger interface
type Logger interface {
	Logf(format string, args ...interface{})
}

// LogFunc type is an adapter to allow the use of ordinary functions as Logger
type LogFunc func(format string, args ...interface{})

// Logf calls f(id)
func (f LogFunc) Logf(format string, args ...interface{}) { f(format, args...) }
