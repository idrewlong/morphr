package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"morphr/internal/config"
	"morphr/internal/meta"
)

var infoCmd = &cobra.Command{
	Use:   "info [file]",
	Short: "Show metadata from a RAW file",
	Long: `Display EXIF and other metadata embedded in a RAW image file.

Example:
  morphr info photo.NEF`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", inputPath)
	}

	ext := strings.ToLower(filepath.Ext(inputPath))
	if !config.SupportedInputFormats[ext] {
		return fmt.Errorf("unsupported format: %s", ext)
	}

	exifData, err := meta.ReadEXIF(inputPath)
	if err != nil {
		return fmt.Errorf("read metadata: %w", err)
	}

	fmt.Printf("File:          %s\n", filepath.Base(inputPath))
	fmt.Printf("Format:        %s\n", strings.ToUpper(strings.TrimPrefix(ext, ".")))

	if exifData.CameraMake != "" {
		fmt.Printf("Camera:        %s %s\n",
			strings.TrimSpace(exifData.CameraMake),
			strings.TrimSpace(exifData.CameraModel))
	}
	if !exifData.DateTime.IsZero() {
		fmt.Printf("Date:          %s\n", exifData.DateTime.Format("2006-01-02 15:04:05"))
	}
	if exifData.ISO > 0 {
		fmt.Printf("ISO:           %d\n", exifData.ISO)
	}
	if exifData.Aperture > 0 {
		fmt.Printf("Aperture:      f/%.1f\n", exifData.Aperture)
	}
	if exifData.ShutterSpeed != "" {
		fmt.Printf("Shutter:       %s s\n", exifData.ShutterSpeed)
	}
	if exifData.FocalLength > 0 {
		fmt.Printf("Focal Length:  %.0f mm\n", exifData.FocalLength)
	}
	if exifData.LensModel != "" {
		fmt.Printf("Lens:          %s\n", exifData.LensModel)
	}
	if exifData.Orientation > 0 {
		fmt.Printf("Orientation:   %d\n", exifData.Orientation)
	}
	if exifData.Software != "" {
		fmt.Printf("Software:      %s\n", exifData.Software)
	}
	if exifData.Copyright != "" {
		fmt.Printf("Copyright:     %s\n", exifData.Copyright)
	}
	if exifData.GPSLatitude != 0 || exifData.GPSLongitude != 0 {
		fmt.Printf("GPS:           %.6f, %.6f\n", exifData.GPSLatitude, exifData.GPSLongitude)
	}

	xmpData, err := meta.ReadXMPSidecar(inputPath)
	if err == nil && xmpData != nil {
		if xmpData.Rating > 0 {
			fmt.Printf("Rating:        %d/5\n", xmpData.Rating)
		}
		if xmpData.Label != "" {
			fmt.Printf("Label:         %s\n", xmpData.Label)
		}
	}

	return nil
}
