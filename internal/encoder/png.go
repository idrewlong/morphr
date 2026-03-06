package encoder

import (
	"image"
	"image/png"
	"io"
)

type PNGEncoder struct{}

func (e *PNGEncoder) Encode(w io.Writer, img image.Image, opts Options) error {
	enc := &png.Encoder{CompressionLevel: png.DefaultCompression}
	return enc.Encode(w, img)
}

func (e *PNGEncoder) Format() string    { return "png" }
func (e *PNGEncoder) Extension() string { return ".png" }
