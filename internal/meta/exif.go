package meta

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

type EXIFData struct {
	CameraMake   string
	CameraModel  string
	DateTime     time.Time
	ISO          int
	Aperture     float64
	ShutterSpeed string
	FocalLength  float64
	LensModel    string
	Orientation  int
	Software     string
	Copyright    string
	ImageWidth   int
	ImageHeight  int
	GPSLatitude  float64
	GPSLongitude float64
}

func ReadEXIF(path string) (*EXIFData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode EXIF: %w", err)
	}

	data := &EXIFData{}

	if tag, err := x.Get(exif.Make); err == nil {
		data.CameraMake, _ = tag.StringVal()
	}
	if tag, err := x.Get(exif.Model); err == nil {
		data.CameraModel, _ = tag.StringVal()
	}
	if tag, err := x.Get(exif.DateTime); err == nil {
		if s, err := tag.StringVal(); err == nil {
			data.DateTime, _ = time.Parse("2006:01:02 15:04:05", s)
		}
	}
	if tag, err := x.Get(exif.ISOSpeedRatings); err == nil {
		data.ISO, _ = tag.Int(0)
	}
	if tag, err := x.Get(exif.FNumber); err == nil {
		num, den, _ := tag.Rat2(0)
		if den != 0 {
			data.Aperture = float64(num) / float64(den)
		}
	}
	if tag, err := x.Get(exif.ExposureTime); err == nil {
		num, den, _ := tag.Rat2(0)
		if den != 0 {
			if num == 1 {
				data.ShutterSpeed = fmt.Sprintf("1/%d", den)
			} else {
				data.ShutterSpeed = fmt.Sprintf("%d/%d", num, den)
			}
		}
	}
	if tag, err := x.Get(exif.FocalLength); err == nil {
		num, den, _ := tag.Rat2(0)
		if den != 0 {
			data.FocalLength = float64(num) / float64(den)
		}
	}
	if tag, err := x.Get(exif.Orientation); err == nil {
		data.Orientation, _ = tag.Int(0)
	}
	if tag, err := x.Get(exif.Software); err == nil {
		data.Software, _ = tag.StringVal()
	}
	if tag, err := x.Get(exif.Copyright); err == nil {
		data.Copyright, _ = tag.StringVal()
	}

	lat, lon, err := x.LatLong()
	if err == nil {
		data.GPSLatitude = lat
		data.GPSLongitude = lon
	}

	return data, nil
}

// PreserveEXIF copies EXIF/IPTC/XMP metadata from src to dst using exiftool.
func PreserveEXIF(srcPath, dstPath string) error {
	exiftoolBin, err := exec.LookPath("exiftool")
	if err != nil {
		return fmt.Errorf("EXIF preservation requires exiftool (install with: brew install exiftool)")
	}

	cmd := exec.Command(exiftoolBin, "-overwrite_original", "-TagsFromFile", srcPath, dstPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("exiftool failed: %s: %w", string(output), err)
	}
	return nil
}
