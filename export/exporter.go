package export

import (
	"io"
)

// Exporter defines interface to export statements list to writer with specified format
type Exporter interface {
	Export(w io.Writer, f Format) error
}
