package encoder

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
)

// WebPEncoder uses the cwebp command-line tool for encoding.
// Install with: brew install webp (macOS), apt install webp (Linux).
type WebPEncoder struct{}

func (e *WebPEncoder) Encode(w io.Writer, img image.Image, opts Options) error {
	cwebp, err := exec.LookPath("cwebp")
	if err != nil {
		return fmt.Errorf("WebP encoding requires cwebp — install with: brew install webp")
	}

	tmpIn, err := os.CreateTemp("", "morphr-*.png")
	if err != nil {
		return fmt.Errorf("create temp input: %w", err)
	}
	defer os.Remove(tmpIn.Name())

	if err := png.Encode(tmpIn, img); err != nil {
		tmpIn.Close()
		return fmt.Errorf("encode temp PNG: %w", err)
	}
	tmpIn.Close()

	tmpOut, err := os.CreateTemp("", "morphr-*.webp")
	if err != nil {
		return fmt.Errorf("create temp output: %w", err)
	}
	defer os.Remove(tmpOut.Name())
	tmpOut.Close()

	quality := opts.Quality
	if quality <= 0 || quality > 100 {
		quality = 90
	}

	cmd := exec.Command(cwebp, "-q", fmt.Sprintf("%d", quality), tmpIn.Name(), "-o", tmpOut.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cwebp failed: %s: %w", string(output), err)
	}

	data, err := os.ReadFile(tmpOut.Name())
	if err != nil {
		return fmt.Errorf("read WebP output: %w", err)
	}

	_, err = w.Write(data)
	return err
}

func (e *WebPEncoder) Format() string    { return "webp" }
func (e *WebPEncoder) Extension() string { return ".webp" }
