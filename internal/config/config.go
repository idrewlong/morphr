package config

import (
	"runtime"
	"strings"
)

type Config struct {
	InputPath    string
	OutputPath   string
	Format       string
	Quality      int
	Overwrite    bool
	MaxWidth     int
	MaxHeight    int
	Scale        float64
	PreserveExif bool
	AutoRotate   bool
	ColorSpace   string
	Workers      int
	Recursive    bool
	Naming       string
	DryRun       bool
	Verbose      bool
	Silent       bool
}

func Default() *Config {
	return &Config{
		Format:       "jpeg",
		Quality:      90,
		PreserveExif: true,
		AutoRotate:   true,
		ColorSpace:   "srgb",
		Workers:      runtime.NumCPU(),
		Naming:       "{name}",
	}
}

func (c *Config) OutputExtension() string {
	if ext, ok := SupportedOutputFormats[strings.ToLower(c.Format)]; ok {
		return ext
	}
	return ".jpg"
}

var SupportedInputFormats = map[string]bool{
	".cr2": true, ".cr3": true, ".nef": true, ".arw": true,
	".dng": true, ".raf": true, ".orf": true, ".rw2": true,
	".pef": true, ".srw": true, ".raw": true, ".3fr": true,
	".ari": true, ".bay": true, ".crw": true, ".dcr": true,
	".erf": true, ".fff": true, ".iiq": true, ".k25": true,
	".kdc": true, ".mef": true, ".mos": true, ".mrw": true,
	".nrw": true, ".ptx": true, ".r3d": true, ".rwl": true,
	".sr2": true, ".srf": true, ".x3f": true,
}

var SupportedOutputFormats = map[string]string{
	"jpeg": ".jpg",
	"jpg":  ".jpg",
	"png":  ".png",
	"tiff": ".tif",
	"tif":  ".tif",
	"webp": ".webp",
	"avif": ".avif",
}
