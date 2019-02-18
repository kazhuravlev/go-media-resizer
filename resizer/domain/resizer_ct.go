package domain

import "bytes"

var (
	_ Resizer = &ContentTypeResizer{}
)

// ContentTypeResizer use one of Mapping resizers depends on content type. If
// content type has no resizer, DefaultResizer will be used.
type ContentTypeResizer struct {
	Mapping        map[string]Resizer
	DefaultResizer Resizer
}

// Resize choose resizer for handle image
func (c *ContentTypeResizer) Resize(ct string, inBuf, outBuf *bytes.Buffer, w int, h int) error {
	resizer, exists := c.Mapping[ct]
	if !exists {
		resizer = c.DefaultResizer
	}

	return resizer.Resize(ct, inBuf, outBuf, w, h)
}
