package decoder

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
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
	meta := &Metadata{
		Width:       bounds.Dx(),
		Height:      bounds.Dy(),
		Orientation: 1,
	}

	// Read EXIF to populate orientation and camera metadata.
	// Open a second handle so image.Decode's read position doesn't interfere.
	if ef, err := os.Open(path); err == nil {
		defer ef.Close()
		if x, err := exif.Decode(ef); err == nil {
			if tag, err := x.Get(exif.Orientation); err == nil {
				if o, err := tag.Int(0); err == nil {
					meta.Orientation = o
				}
			}
			if tag, err := x.Get(exif.Make); err == nil {
				meta.CameraMake, _ = tag.StringVal()
			}
			if tag, err := x.Get(exif.Model); err == nil {
				meta.CameraModel, _ = tag.StringVal()
			}
			if tag, err := x.Get(exif.ISOSpeedRatings); err == nil {
				if iso, err := tag.Int(0); err == nil {
					meta.ISO = iso
				}
			}
			if t, err := x.DateTime(); err == nil {
				meta.DateTime = t
			}
		}
	}

	return &Result{Image: img, Meta: meta}, nil
}
