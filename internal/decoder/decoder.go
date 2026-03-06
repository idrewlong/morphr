package decoder

import (
	"fmt"
	"image"
	"path/filepath"
	"strings"
	"time"
)

type Metadata struct {
	Width        int
	Height       int
	Orientation  int
	CameraMake   string
	CameraModel  string
	DateTime     time.Time
	ISO          int
	Aperture     float64
	ShutterSpeed string
	FocalLength  float64
	LensModel    string
	Software     string
	Copyright    string
}

type Result struct {
	Image image.Image
	Meta  *Metadata
}

type Decoder interface {
	Decode(path string) (*Result, error)
	Available() bool
	Name() string
}

// Decode tries all registered decoders in priority order until one succeeds.
func Decode(path string) (*Result, error) {
	decoders := []Decoder{
		NewDNGDecoder(),
		NewDcrawDecoder(),
	}

	ext := strings.ToLower(filepath.Ext(path))
	var lastErr error

	for _, d := range decoders {
		if !d.Available() {
			continue
		}
		result, err := d.Decode(path)
		if err != nil {
			lastErr = err
			continue
		}
		return result, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to decode %s: %w", filepath.Base(path), lastErr)
	}
	return nil, fmt.Errorf("no decoder available for %s format — install dcraw or LibRaw", ext)
}
