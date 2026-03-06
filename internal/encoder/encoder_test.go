package encoder

import (
	"bytes"
	"image"
	"image/color"
	"os/exec"
	"testing"
)

func syntheticImage(w, h int) image.Image {
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
	return img
}

func TestJPEGEncoder(t *testing.T) {
	img := syntheticImage(100, 80)
	enc := &JPEGEncoder{}

	tests := []struct {
		name    string
		quality int
	}{
		{"default quality", 0},
		{"low quality", 10},
		{"medium quality", 50},
		{"high quality", 95},
		{"max quality", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := enc.Encode(&buf, img, Options{Quality: tt.quality})
			if err != nil {
				t.Fatalf("encode failed: %v", err)
			}
			if buf.Len() == 0 {
				t.Fatal("output is empty")
			}
			// JPEG magic bytes: FF D8
			if buf.Bytes()[0] != 0xFF || buf.Bytes()[1] != 0xD8 {
				t.Fatal("output is not valid JPEG")
			}
		})
	}
}

func TestJPEGQualityAffectsSize(t *testing.T) {
	img := syntheticImage(200, 150)
	enc := &JPEGEncoder{}

	var lowBuf, highBuf bytes.Buffer
	if err := enc.Encode(&lowBuf, img, Options{Quality: 10}); err != nil {
		t.Fatalf("low quality encode: %v", err)
	}
	if err := enc.Encode(&highBuf, img, Options{Quality: 95}); err != nil {
		t.Fatalf("high quality encode: %v", err)
	}

	if lowBuf.Len() >= highBuf.Len() {
		t.Errorf("expected low quality (%d bytes) < high quality (%d bytes)", lowBuf.Len(), highBuf.Len())
	}
}

func TestPNGEncoder(t *testing.T) {
	img := syntheticImage(100, 80)
	enc := &PNGEncoder{}

	var buf bytes.Buffer
	err := enc.Encode(&buf, img, Options{})
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("output is empty")
	}
	// PNG magic bytes: 89 50 4E 47
	magic := buf.Bytes()[:4]
	if magic[0] != 0x89 || magic[1] != 0x50 || magic[2] != 0x4E || magic[3] != 0x47 {
		t.Fatal("output is not valid PNG")
	}
}

func TestTIFFEncoder(t *testing.T) {
	img := syntheticImage(100, 80)
	enc := &TIFFEncoder{}

	var buf bytes.Buffer
	err := enc.Encode(&buf, img, Options{})
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("output is empty")
	}
	// TIFF magic bytes: 49 49 (little-endian) or 4D 4D (big-endian)
	b := buf.Bytes()[:2]
	if !((b[0] == 0x49 && b[1] == 0x49) || (b[0] == 0x4D && b[1] == 0x4D)) {
		t.Fatal("output is not valid TIFF")
	}
}

func TestWebPEncoder(t *testing.T) {
	if _, err := exec.LookPath("cwebp"); err != nil {
		t.Skip("cwebp not installed, skipping WebP test")
	}

	img := syntheticImage(100, 80)
	enc := &WebPEncoder{}

	var buf bytes.Buffer
	err := enc.Encode(&buf, img, Options{Quality: 80})
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("output is empty")
	}
	// WebP starts with RIFF header
	if string(buf.Bytes()[:4]) != "RIFF" {
		t.Fatal("output is not valid WebP")
	}
}

func TestAVIFEncoder(t *testing.T) {
	if _, err := exec.LookPath("avifenc"); err != nil {
		t.Skip("avifenc not installed, skipping AVIF test")
	}

	img := syntheticImage(100, 80)
	enc := &AVIFEncoder{}

	var buf bytes.Buffer
	err := enc.Encode(&buf, img, Options{Quality: 80})
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("output is empty")
	}
}

func TestGetEncoder(t *testing.T) {
	tests := []struct {
		format  string
		wantFmt string
		wantExt string
		wantErr bool
	}{
		{"jpeg", "jpeg", ".jpg", false},
		{"jpg", "jpeg", ".jpg", false},
		{"png", "png", ".png", false},
		{"tiff", "tiff", ".tif", false},
		{"tif", "tiff", ".tif", false},
		{"webp", "webp", ".webp", false},
		{"avif", "avif", ".avif", false},
		{"JPEG", "jpeg", ".jpg", false},
		{"bmp", "", "", true},
		{"gif", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			enc, err := Get(tt.format)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for format %q", tt.format)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if enc.Format() != tt.wantFmt {
				t.Errorf("format = %q, want %q", enc.Format(), tt.wantFmt)
			}
			if enc.Extension() != tt.wantExt {
				t.Errorf("extension = %q, want %q", enc.Extension(), tt.wantExt)
			}
		})
	}
}

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()
	if len(formats) < 5 {
		t.Errorf("expected at least 5 formats, got %d", len(formats))
	}

	has := map[string]bool{}
	for _, f := range formats {
		has[f] = true
	}
	for _, want := range []string{"jpeg", "png", "tiff", "webp", "avif"} {
		if !has[want] {
			t.Errorf("missing format: %s", want)
		}
	}
}

func BenchmarkJPEG(b *testing.B) {
	img := syntheticImage(800, 600)
	enc := &JPEGEncoder{}
	opts := Options{Quality: 85}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc.Encode(&buf, img, opts)
	}
}

func BenchmarkPNG(b *testing.B) {
	img := syntheticImage(800, 600)
	enc := &PNGEncoder{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc.Encode(&buf, img, Options{})
	}
}

func BenchmarkTIFF(b *testing.B) {
	img := syntheticImage(800, 600)
	enc := &TIFFEncoder{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc.Encode(&buf, img, Options{})
	}
}
