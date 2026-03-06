package processor

import (
	"image"
	"image/color"
	"testing"
)

func syntheticImage(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: 100,
				A: 255,
			})
		}
	}
	return img
}

func TestResizeScale(t *testing.T) {
	img := syntheticImage(400, 300)

	tests := []struct {
		name   string
		scale  float64
		wantW  int
		wantH  int
	}{
		{"50%", 50, 200, 150},
		{"25%", 25, 100, 75},
		{"200%", 200, 800, 600},
		{"no scale (100%)", 100, 400, 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Resize(img, 0, 0, tt.scale)
			bounds := result.Bounds()

			if tt.scale == 100 {
				if bounds.Dx() != 400 || bounds.Dy() != 300 {
					t.Errorf("100%% scale should return original, got %dx%d", bounds.Dx(), bounds.Dy())
				}
				return
			}

			if bounds.Dx() != tt.wantW || bounds.Dy() != tt.wantH {
				t.Errorf("got %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), tt.wantW, tt.wantH)
			}
		})
	}
}

func TestResizeMaxWidth(t *testing.T) {
	img := syntheticImage(1000, 500)

	result := Resize(img, 500, 0, 0)
	bounds := result.Bounds()

	if bounds.Dx() > 500 {
		t.Errorf("width should be <= 500, got %d", bounds.Dx())
	}
	// Aspect ratio: 1000:500 = 2:1, so 500 wide → 250 tall
	if bounds.Dy() != 250 {
		t.Errorf("height should be 250, got %d", bounds.Dy())
	}
}

func TestResizeMaxHeight(t *testing.T) {
	img := syntheticImage(800, 1200)

	result := Resize(img, 0, 600, 0)
	bounds := result.Bounds()

	if bounds.Dy() > 600 {
		t.Errorf("height should be <= 600, got %d", bounds.Dy())
	}
	// Aspect ratio: 800:1200 = 2:3, so 600 tall → 400 wide
	if bounds.Dx() != 400 {
		t.Errorf("width should be 400, got %d", bounds.Dx())
	}
}

func TestResizeFitBox(t *testing.T) {
	img := syntheticImage(2000, 1000)

	result := Resize(img, 800, 600, 0)
	bounds := result.Bounds()

	if bounds.Dx() > 800 || bounds.Dy() > 600 {
		t.Errorf("should fit in 800x600 box, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestResizeNoUpscale(t *testing.T) {
	img := syntheticImage(200, 100)

	result := Resize(img, 800, 600, 0)
	bounds := result.Bounds()

	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("should not upscale, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestResizeNoOp(t *testing.T) {
	img := syntheticImage(400, 300)

	result := Resize(img, 0, 0, 0)
	if result != img {
		t.Error("no-op resize should return original image")
	}
}

func TestAutoRotateIdentity(t *testing.T) {
	img := syntheticImage(100, 50)

	result := AutoRotate(img, 1)
	if result != img {
		t.Error("orientation 1 should return original image")
	}

	result = AutoRotate(img, 0)
	if result != img {
		t.Error("orientation 0 should return original image")
	}
}

func TestAutoRotateDimensions(t *testing.T) {
	img := syntheticImage(200, 100) // landscape

	tests := []struct {
		orientation int
		wantW       int
		wantH       int
	}{
		{1, 200, 100},
		{2, 200, 100}, // horizontal flip
		{3, 200, 100}, // 180 rotation
		{4, 200, 100}, // vertical flip
		{5, 100, 200}, // transpose (swaps dims)
		{6, 100, 200}, // 270 rotation (swaps dims)
		{7, 100, 200}, // transverse (swaps dims)
		{8, 100, 200}, // 90 rotation (swaps dims)
	}

	for _, tt := range tests {
		t.Run("orientation_"+string(rune('0'+tt.orientation)), func(t *testing.T) {
			result := AutoRotate(img, tt.orientation)
			bounds := result.Bounds()
			if bounds.Dx() != tt.wantW || bounds.Dy() != tt.wantH {
				t.Errorf("orientation %d: got %dx%d, want %dx%d",
					tt.orientation, bounds.Dx(), bounds.Dy(), tt.wantW, tt.wantH)
			}
		})
	}
}

func TestConvertColorSpaceNoOp(t *testing.T) {
	img := syntheticImage(50, 50)

	result := ConvertColorSpace(img, "srgb")
	if result != img {
		t.Error("srgb target should return original")
	}

	result = ConvertColorSpace(img, "")
	if result != img {
		t.Error("empty target should return original")
	}
}

func TestConvertColorSpaceAdobeRGB(t *testing.T) {
	img := syntheticImage(50, 50)
	result := ConvertColorSpace(img, "adobergb")

	bounds := result.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("dimensions changed: got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Verify output is a different image (pixels should differ)
	origR, origG, origB, _ := img.At(25, 25).RGBA()
	newR, newG, newB, _ := result.At(25, 25).RGBA()

	if origR == newR && origG == newG && origB == newB {
		t.Error("adobergb conversion should change pixel values")
	}
}

func TestConvertColorSpaceProPhoto(t *testing.T) {
	img := syntheticImage(50, 50)
	result := ConvertColorSpace(img, "prophoto")

	bounds := result.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("dimensions changed: got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func BenchmarkResize(b *testing.B) {
	img := syntheticImage(2000, 1500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Resize(img, 800, 600, 0)
	}
}

func BenchmarkAutoRotate(b *testing.B) {
	img := syntheticImage(2000, 1500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AutoRotate(img, 6)
	}
}

func BenchmarkColorConvert(b *testing.B) {
	img := syntheticImage(800, 600)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConvertColorSpace(img, "adobergb")
	}
}
