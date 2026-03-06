package config

import (
	"runtime"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	if cfg.Format != "jpeg" {
		t.Errorf("Format = %q, want %q", cfg.Format, "jpeg")
	}
	if cfg.Quality != 90 {
		t.Errorf("Quality = %d, want 90", cfg.Quality)
	}
	if !cfg.PreserveExif {
		t.Error("PreserveExif should default to true")
	}
	if !cfg.AutoRotate {
		t.Error("AutoRotate should default to true")
	}
	if cfg.ColorSpace != "srgb" {
		t.Errorf("ColorSpace = %q, want %q", cfg.ColorSpace, "srgb")
	}
	if cfg.Workers != runtime.NumCPU() {
		t.Errorf("Workers = %d, want %d", cfg.Workers, runtime.NumCPU())
	}
	if cfg.Naming != "{name}" {
		t.Errorf("Naming = %q, want %q", cfg.Naming, "{name}")
	}
	if cfg.Overwrite {
		t.Error("Overwrite should default to false")
	}
	if cfg.Recursive {
		t.Error("Recursive should default to false")
	}
	if cfg.DryRun {
		t.Error("DryRun should default to false")
	}
}

func TestOutputExtension(t *testing.T) {
	tests := []struct {
		format  string
		wantExt string
	}{
		{"jpeg", ".jpg"},
		{"jpg", ".jpg"},
		{"png", ".png"},
		{"tiff", ".tif"},
		{"tif", ".tif"},
		{"webp", ".webp"},
		{"avif", ".avif"},
		{"unknown", ".jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			cfg := &Config{Format: tt.format}
			got := cfg.OutputExtension()
			if got != tt.wantExt {
				t.Errorf("OutputExtension(%q) = %q, want %q", tt.format, got, tt.wantExt)
			}
		})
	}
}

func TestSupportedInputFormats(t *testing.T) {
	required := []string{
		".cr2", ".cr3", ".nef", ".arw", ".dng",
		".raf", ".orf", ".rw2", ".pef", ".srw",
	}

	for _, ext := range required {
		if !SupportedInputFormats[ext] {
			t.Errorf("missing required input format: %s", ext)
		}
	}

	unsupported := []string{".jpg", ".png", ".tiff", ".bmp", ".gif"}
	for _, ext := range unsupported {
		if SupportedInputFormats[ext] {
			t.Errorf("should not support %s as input", ext)
		}
	}
}

func TestSupportedOutputFormats(t *testing.T) {
	required := []string{"jpeg", "jpg", "png", "tiff", "tif", "webp", "avif"}

	for _, fmt := range required {
		if _, ok := SupportedOutputFormats[fmt]; !ok {
			t.Errorf("missing required output format: %s", fmt)
		}
	}
}
