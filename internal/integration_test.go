package internal

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/tiff"

	"morphr/internal/decoder"
	"morphr/internal/encoder"
	"morphr/internal/processor"
)

func createTestTIFF(t *testing.T, dir string, w, h int) string {
	t.Helper()
	path := filepath.Join(dir, "test.tiff")

	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8((x * 255) / w),
				G: uint8((y * 255) / h),
				B: 128,
				A: 255,
			})
		}
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create tiff: %v", err)
	}
	defer f.Close()

	if err := tiff.Encode(f, img, nil); err != nil {
		t.Fatalf("encode tiff: %v", err)
	}

	return path
}

func TestEndToEndTIFFToJPEG(t *testing.T) {
	tmpDir := t.TempDir()
	tiffPath := createTestTIFF(t, tmpDir, 200, 150)

	result, err := decoder.Decode(tiffPath)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	img := result.Image
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 150 {
		t.Fatalf("decoded dimensions = %dx%d, want 200x150", bounds.Dx(), bounds.Dy())
	}

	img = processor.Resize(img, 100, 0, 0)
	bounds = img.Bounds()
	if bounds.Dx() != 100 {
		t.Errorf("resized width = %d, want 100", bounds.Dx())
	}

	enc, err := encoder.Get("jpeg")
	if err != nil {
		t.Fatalf("get encoder: %v", err)
	}

	var buf bytes.Buffer
	if err := enc.Encode(&buf, img, encoder.Options{Quality: 85}); err != nil {
		t.Fatalf("encode: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("JPEG output is empty")
	}
	if buf.Bytes()[0] != 0xFF || buf.Bytes()[1] != 0xD8 {
		t.Fatal("output is not valid JPEG")
	}
}

func TestEndToEndTIFFToPNG(t *testing.T) {
	tmpDir := t.TempDir()
	tiffPath := createTestTIFF(t, tmpDir, 120, 90)

	result, err := decoder.Decode(tiffPath)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	img := processor.AutoRotate(result.Image, 3)
	bounds := img.Bounds()
	if bounds.Dx() != 120 || bounds.Dy() != 90 {
		t.Errorf("rotated 180 should keep dimensions, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	enc, err := encoder.Get("png")
	if err != nil {
		t.Fatalf("get encoder: %v", err)
	}

	var buf bytes.Buffer
	if err := enc.Encode(&buf, img, encoder.Options{}); err != nil {
		t.Fatalf("encode: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("PNG output is empty")
	}
	magic := buf.Bytes()[:4]
	if magic[0] != 0x89 || magic[1] != 0x50 || magic[2] != 0x4E || magic[3] != 0x47 {
		t.Fatal("output is not valid PNG")
	}
}

func TestEndToEndTIFFToTIFF(t *testing.T) {
	tmpDir := t.TempDir()
	tiffPath := createTestTIFF(t, tmpDir, 300, 200)

	result, err := decoder.Decode(tiffPath)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	img := processor.Resize(result.Image, 0, 0, 50)
	bounds := img.Bounds()
	if bounds.Dx() != 150 || bounds.Dy() != 100 {
		t.Errorf("50%% resize: got %dx%d, want 150x100", bounds.Dx(), bounds.Dy())
	}

	enc, err := encoder.Get("tiff")
	if err != nil {
		t.Fatalf("get encoder: %v", err)
	}

	var buf bytes.Buffer
	if err := enc.Encode(&buf, img, encoder.Options{}); err != nil {
		t.Fatalf("encode: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("TIFF output is empty")
	}
}

func TestEndToEndWithColorConversion(t *testing.T) {
	tmpDir := t.TempDir()
	tiffPath := createTestTIFF(t, tmpDir, 100, 100)

	result, err := decoder.Decode(tiffPath)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	img := processor.ConvertColorSpace(result.Image, "adobergb")
	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("color conversion changed dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}

	enc, err := encoder.Get("jpeg")
	if err != nil {
		t.Fatalf("get encoder: %v", err)
	}

	var buf bytes.Buffer
	if err := enc.Encode(&buf, img, encoder.Options{Quality: 90}); err != nil {
		t.Fatalf("encode: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("output is empty")
	}
}

func TestEndToEndFullPipeline(t *testing.T) {
	tmpDir := t.TempDir()
	tiffPath := createTestTIFF(t, tmpDir, 800, 600)

	result, err := decoder.Decode(tiffPath)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	img := result.Image
	img = processor.AutoRotate(img, 6)
	img = processor.Resize(img, 400, 300, 0)
	img = processor.ConvertColorSpace(img, "prophoto")

	for _, format := range []string{"jpeg", "png", "tiff"} {
		t.Run(format, func(t *testing.T) {
			enc, err := encoder.Get(format)
			if err != nil {
				t.Fatalf("get encoder: %v", err)
			}

			outPath := filepath.Join(tmpDir, "output"+enc.Extension())
			f, err := os.Create(outPath)
			if err != nil {
				t.Fatalf("create output: %v", err)
			}
			defer f.Close()

			if err := enc.Encode(f, img, encoder.Options{Quality: 85}); err != nil {
				t.Fatalf("encode %s: %v", format, err)
			}

			info, err := os.Stat(outPath)
			if err != nil {
				t.Fatalf("stat output: %v", err)
			}
			if info.Size() == 0 {
				t.Errorf("%s output file is empty", format)
			}
		})
	}
}
