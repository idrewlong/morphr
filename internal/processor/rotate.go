package processor

import (
	"image"

	"github.com/disintegration/imaging"
)

// AutoRotate applies the correct transformation based on the EXIF orientation tag.
// Orientation values 1–8 follow the EXIF specification.
func AutoRotate(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.Transpose(img)
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.Transverse(img)
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}
