package decoder

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/tiff"
)

// DNGDecoder handles DNG and standard TIFF files using pure Go decoding.
// Works well for linear (pre-demosaiced) DNG files. Bayer-pattern DNGs
// will fall through to the dcraw decoder.
type DNGDecoder struct{}

func NewDNGDecoder() *DNGDecoder {
	return &DNGDecoder{}
}

func (d *DNGDecoder) Available() bool {
	return true
}

func (d *DNGDecoder) Name() string {
	return "dng"
}

func (d *DNGDecoder) Decode(path string) (*Result, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".dng" && ext != ".tif" && ext != ".tiff" {
		return nil, fmt.Errorf("DNG decoder only handles .dng/.tif/.tiff files, got %s", ext)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s as %s: %w", filepath.Base(path), format, err)
	}

	bounds := img.Bounds()
	return &Result{
		Image: img,
		Meta: &Metadata{
			Width:  bounds.Dx(),
			Height: bounds.Dy(),
		},
	}, nil
}
