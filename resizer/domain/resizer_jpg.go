package domain

import (
	"bytes"
	"github.com/disintegration/imaging"
	"image"
	"image/jpeg"
	"log"
)

var (
	_ Resizer = &JPEG{}
)

// JPEG implements Resizer interface for jpeg images
type JPEG struct{}

// Resize is a function, that resize input image
func (c *JPEG) Resize(ct string, inBuf, outBuf *bytes.Buffer, w int, h int) error {
	img, _, err := image.Decode(inBuf)
	if err != nil {
		log.Println(err)
		return err
	}

	img = imaging.Resize(img, w, h, imaging.Lanczos)

	if err := jpeg.Encode(outBuf, img, &jpeg.Options{Quality: 95}); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
