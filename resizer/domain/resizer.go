package domain

import (
	"bytes"
)

// Resizer is an interface for resize image
type Resizer interface {
	Resize(ct string, inBuf, outBuf *bytes.Buffer, w int, h int) error
}
