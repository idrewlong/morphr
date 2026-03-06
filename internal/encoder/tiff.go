package encoder

import (
	"image"
	"io"

	"golang.org/x/image/tiff"
)

type TIFFEncoder struct{}

func (e *TIFFEncoder) Encode(w io.Writer, img image.Image, opts Options) error {
	return tiff.Encode(w, img, &tiff.Options{
		Compression: tiff.Deflate,
	})
}

func (e *TIFFEncoder) Format() string    { return "tiff" }
func (e *TIFFEncoder) Extension() string { return ".tif" }
