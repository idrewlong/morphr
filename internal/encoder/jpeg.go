package encoder

import (
	"image"
	"image/jpeg"
	"io"
)

type JPEGEncoder struct{}

func (e *JPEGEncoder) Encode(w io.Writer, img image.Image, opts Options) error {
	quality := opts.Quality
	if quality <= 0 || quality > 100 {
		quality = 90
	}
	return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
}

func (e *JPEGEncoder) Format() string    { return "jpeg" }
func (e *JPEGEncoder) Extension() string { return ".jpg" }
