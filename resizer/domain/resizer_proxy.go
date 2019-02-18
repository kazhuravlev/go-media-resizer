package domain

import (
	"bytes"
	"io"
)

var (
	_ Resizer = &Proxy{}
)

// Proxy is a dummy Resizer - just proxy bytes to result buffer
type Proxy struct{}

// Resize just copy bytes to result buf
func (c *Proxy) Resize(ct string, inBuf, outBuf *bytes.Buffer, w int, h int) error {
	if _, err := io.Copy(outBuf, inBuf); err != nil {
		return err
	}

	return nil
}
