package processor

import (
	"image"
	"image/color"
	"math"
)

type colorMatrix [3][3]float64

// Approximate conversion matrices from sRGB to other working spaces.
// Full ICC profile support is a future enhancement.
var (
	sRGBToAdobeRGB = colorMatrix{
		{0.7152, 0.2848, 0.0000},
		{0.0000, 1.0000, 0.0000},
		{0.0000, 0.0413, 0.9587},
	}
	sRGBToProPhoto = colorMatrix{
		{0.5293, 0.3300, 0.1407},
		{0.0980, 0.8734, 0.0286},
		{0.0168, 0.0184, 0.9648},
	}
)

// ConvertColorSpace transforms an image from sRGB to the target color space.
// Supported targets: "srgb" (no-op), "adobergb", "prophoto".
func ConvertColorSpace(img image.Image, target string) image.Image {
	switch target {
	case "adobergb":
		return applyMatrix(img, sRGBToAdobeRGB)
	case "prophoto":
		return applyMatrix(img, sRGBToProPhoto)
	default:
		return img
	}
}

func linearize(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func delinearize(v float64) float64 {
	if v <= 0.0031308 {
		return v * 12.92
	}
	return 1.055*math.Pow(v, 1.0/2.4) - 0.055
}

func clamp(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}

func applyMatrix(img image.Image, m colorMatrix) image.Image {
	bounds := img.Bounds()
	result := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			rf := linearize(float64(r) / 65535.0)
			gf := linearize(float64(g) / 65535.0)
			bf := linearize(float64(b) / 65535.0)

			nr := m[0][0]*rf + m[0][1]*gf + m[0][2]*bf
			ng := m[1][0]*rf + m[1][1]*gf + m[1][2]*bf
			nb := m[2][0]*rf + m[2][1]*gf + m[2][2]*bf

			result.SetNRGBA(x, y, color.NRGBA{
				R: uint8(clamp(delinearize(nr)) * 255),
				G: uint8(clamp(delinearize(ng)) * 255),
				B: uint8(clamp(delinearize(nb)) * 255),
				A: uint8(a >> 8),
			})
		}
	}

	return result
}
