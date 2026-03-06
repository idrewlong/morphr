package encoder

import (
	"fmt"
	"image"
	"io"
	"strings"
)

type Options struct {
	Quality int
}

type Encoder interface {
	Encode(w io.Writer, img image.Image, opts Options) error
	Format() string
	Extension() string
}

var encoders = map[string]Encoder{}

func Register(e Encoder) {
	encoders[strings.ToLower(e.Format())] = e
}

func Get(format string) (Encoder, error) {
	format = strings.ToLower(format)
	switch format {
	case "jpg":
		format = "jpeg"
	case "tif":
		format = "tiff"
	}

	e, ok := encoders[format]
	if !ok {
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
	return e, nil
}

func SupportedFormats() []string {
	formats := make([]string, 0, len(encoders))
	for f := range encoders {
		formats = append(formats, f)
	}
	return formats
}

func init() {
	Register(&JPEGEncoder{})
	Register(&PNGEncoder{})
	Register(&TIFFEncoder{})
	Register(&WebPEncoder{})
	Register(&AVIFEncoder{})
}
