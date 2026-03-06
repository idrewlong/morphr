package processor

import (
	"image"

	"github.com/disintegration/imaging"
)

// Resize scales an image according to the given constraints.
// If scale is set (1-100), it takes priority over max dimensions.
// If maxWidth/maxHeight are set, the image is fit into the bounding box
// while preserving aspect ratio. No upscaling is performed.
func Resize(img image.Image, maxWidth, maxHeight int, scale float64) image.Image {
	if scale > 0 && scale != 100 {
		bounds := img.Bounds()
		w := int(float64(bounds.Dx()) * scale / 100)
		h := int(float64(bounds.Dy()) * scale / 100)
		if w < 1 {
			w = 1
		}
		if h < 1 {
			h = 1
		}
		return imaging.Resize(img, w, h, imaging.Lanczos)
	}

	if maxWidth <= 0 && maxHeight <= 0 {
		return img
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	if maxWidth <= 0 {
		maxWidth = w
	}
	if maxHeight <= 0 {
		maxHeight = h
	}

	if w <= maxWidth && h <= maxHeight {
		return img
	}

	return imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)
}
