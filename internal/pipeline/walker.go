package pipeline

import (
	"io/fs"
	"path/filepath"
	"strings"

	"morphr/internal/config"
)

type FileEntry struct {
	Path string
	Ext  string
}

// WalkDirectory discovers RAW files in a directory, optionally recursing into subdirectories.
func WalkDirectory(dir string, recursive bool) ([]FileEntry, error) {
	var files []FileEntry

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if !recursive && path != dir {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if config.SupportedInputFormats[ext] {
			files = append(files, FileEntry{
				Path: path,
				Ext:  ext,
			})
		}

		return nil
	})

	return files, err
}
