package decoder

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/tiff"
)

func TestDNGDecoderRejectsNonDNG(t *testing.T) {
	d := NewDNGDecoder()

	extensions := []string{".cr2", ".nef", ".arw", ".jpg", ".png"}
	for _, ext := range extensions {
		t.Run(ext, func(t *testing.T) {
			_, err := d.Decode("fake" + ext)
			if err == nil {
				t.Errorf("DNG decoder should reject %s files", ext)
			}
		})
	}
}

func TestDNGDecoderAvailable(t *testing.T) {
	d := NewDNGDecoder()
	if !d.Available() {
		t.Error("DNG decoder should always be available")
	}
}

func TestDNGDecoderName(t *testing.T) {
	d := NewDNGDecoder()
	if d.Name() != "dng" {
		t.Errorf("Name() = %q, want %q", d.Name(), "dng")
	}
}

func TestDNGDecoderWithTIFF(t *testing.T) {
	tmpDir := t.TempDir()
	tiffPath := filepath.Join(tmpDir, "test.tiff")

	img := image.NewNRGBA(image.Rect(0, 0, 100, 80))
	for y := 0; y < 80; y++ {
		for x := 0; x < 100; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}

	f, err := os.Create(tiffPath)
	if err != nil {
		t.Fatalf("create tiff: %v", err)
	}
	if err := tiff.Encode(f, img, nil); err != nil {
		f.Close()
		t.Fatalf("encode tiff: %v", err)
	}
	f.Close()

	d := NewDNGDecoder()
	result, err := d.Decode(tiffPath)
	if err != nil {
		t.Fatalf("decode tiff: %v", err)
	}

	bounds := result.Image.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 80 {
		t.Errorf("dimensions = %dx%d, want 100x80", bounds.Dx(), bounds.Dy())
	}
	if result.Meta == nil {
		t.Fatal("metadata should not be nil")
	}
	if result.Meta.Width != 100 || result.Meta.Height != 80 {
		t.Errorf("meta dimensions = %dx%d, want 100x80", result.Meta.Width, result.Meta.Height)
	}
}

func TestDNGDecoderMissingFile(t *testing.T) {
	d := NewDNGDecoder()
	_, err := d.Decode("/nonexistent/test.dng")
	if err == nil {
		t.Error("should fail on missing file")
	}
}

func TestDcrawDecoderName(t *testing.T) {
	d := NewDcrawDecoder()
	if d.Name() != "dcraw" {
		t.Errorf("Name() = %q, want %q", d.Name(), "dcraw")
	}
}

func TestDcrawDecoderMissingFile(t *testing.T) {
	d := NewDcrawDecoder()
	if !d.Available() {
		t.Skip("dcraw not installed, skipping")
	}
	_, err := d.Decode("/nonexistent/photo.cr2")
	if err == nil {
		t.Error("should fail on missing file")
	}
}

func TestDecodeUnsupportedExtension(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a PNG file and try to decode it as RAW
	pngPath := filepath.Join(tmpDir, "test.xyz")
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	f, err := os.Create(pngPath)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	png.Encode(f, img)
	f.Close()

	_, err = Decode(pngPath)
	if err == nil {
		t.Error("should fail on unsupported extension")
	}
}

func TestDecodeTIFF(t *testing.T) {
	tmpDir := t.TempDir()
	tiffPath := filepath.Join(tmpDir, "test.tiff")

	img := image.NewNRGBA(image.Rect(0, 0, 64, 48))
	for y := 0; y < 48; y++ {
		for x := 0; x < 64; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 128, B: 64, A: 255})
		}
	}

	f, err := os.Create(tiffPath)
	if err != nil {
		t.Fatalf("create tiff: %v", err)
	}
	if err := tiff.Encode(f, img, nil); err != nil {
		f.Close()
		t.Fatalf("encode tiff: %v", err)
	}
	f.Close()

	result, err := Decode(tiffPath)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	bounds := result.Image.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 48 {
		t.Errorf("dimensions = %dx%d, want 64x48", bounds.Dx(), bounds.Dy())
	}
}
