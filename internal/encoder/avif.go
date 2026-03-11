package encoder

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
)

// AVIFEncoder uses the avifenc command-line tool for encoding.
// Install with: brew install libavif (macOS), apt install libavif-bin (Linux).
type AVIFEncoder struct{}

func (e *AVIFEncoder) Encode(w io.Writer, img image.Image, opts Options) error {
	avifenc, err := exec.LookPath("avifenc")
	if err != nil {
		return fmt.Errorf("AVIF encoding requires avifenc — install with: brew install libavif")
	}

	tmpIn, err := os.CreateTemp("", "morphr-*.png")
	if err != nil {
		return fmt.Errorf("create temp input: %w", err)
	}
	if err := png.Encode(tmpIn, img); err != nil {
		tmpIn.Close()
		os.Remove(tmpIn.Name())
		return fmt.Errorf("encode temp PNG: %w", err)
	}
	tmpIn.Close()
	defer os.Remove(tmpIn.Name())

	tmpOut, err := os.CreateTemp("", "morphr-*.avif")
	if err != nil {
		return fmt.Errorf("create temp output: %w", err)
	}
	tmpOut.Close()
	defer os.Remove(tmpOut.Name())

	quality := opts.Quality
	if quality <= 0 || quality > 100 {
		quality = 90
	}

	// avifenc uses --min/--max (0=best, 63=worst) — map our 1-100 scale
	avifQuality := 63 - (quality * 63 / 100)
	cmd := exec.Command(avifenc,
		"--min", fmt.Sprintf("%d", avifQuality),
		"--max", fmt.Sprintf("%d", avifQuality),
		"--speed", "6",
		tmpIn.Name(), tmpOut.Name(),
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("avifenc failed: %s: %w", string(output), err)
	}

	data, err := os.ReadFile(tmpOut.Name())
	if err != nil {
		return fmt.Errorf("read AVIF output: %w", err)
	}

	_, err = w.Write(data)
	return err
}

func (e *AVIFEncoder) Format() string    { return "avif" }
func (e *AVIFEncoder) Extension() string { return ".avif" }
