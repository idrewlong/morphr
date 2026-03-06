package decoder

import (
	"bytes"
	"fmt"
	"image"
	"os/exec"
	"path/filepath"

	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/tiff"
)

// DcrawDecoder uses dcraw or dcraw_emu (LibRaw) as a subprocess to decode RAW files.
type DcrawDecoder struct {
	binary string
}

func NewDcrawDecoder() *DcrawDecoder {
	for _, bin := range []string{"dcraw_emu", "dcraw"} {
		if path, err := exec.LookPath(bin); err == nil {
			return &DcrawDecoder{binary: path}
		}
	}
	return &DcrawDecoder{}
}

func (d *DcrawDecoder) Available() bool {
	return d.binary != ""
}

func (d *DcrawDecoder) Name() string {
	return "dcraw"
}

func (d *DcrawDecoder) Decode(path string) (*Result, error) {
	if !d.Available() {
		return nil, fmt.Errorf("dcraw is not installed")
	}

	// -c  write to stdout
	// -T  output TIFF format
	// -w  use camera white balance
	// -q 3  AHD interpolation (highest quality demosaicing)
	args := []string{"-c", "-T", "-w", "-q", "3", path}

	cmd := exec.Command(d.binary, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s failed: %s: %w", filepath.Base(d.binary), stderr.String(), err)
	}

	if stdout.Len() == 0 {
		return nil, fmt.Errorf("%s produced no output for %s", filepath.Base(d.binary), filepath.Base(path))
	}

	img, _, err := image.Decode(&stdout)
	if err != nil {
		return nil, fmt.Errorf("decode %s output: %w", filepath.Base(d.binary), err)
	}

	bounds := img.Bounds()
	return &Result{
		Image: img,
		Meta: &Metadata{
			Width:  bounds.Dx(),
			Height: bounds.Dy(),
		},
	}, nil
}
