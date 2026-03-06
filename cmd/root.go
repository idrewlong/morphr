package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "morphr",
	Short: "Fast RAW image converter",
	Long: `Morphr is a fast, dependency-light command-line tool for converting
camera RAW image files (CR2, NEF, ARW, DNG, RAF, ORF, etc.) into
standard formats like JPEG, PNG, TIFF, and WebP.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = fmt.Sprintf("%s (built %s)", Version, BuildTime)

	rootCmd.PersistentFlags().StringP("format", "f", "jpeg", "Output format: jpeg|png|tiff|webp|avif")
	rootCmd.PersistentFlags().IntP("quality", "q", 90, "JPEG/WebP quality 1-100")
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output file or directory")
	rootCmd.PersistentFlags().Bool("overwrite", false, "Overwrite existing files")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose logging")
	rootCmd.PersistentFlags().Bool("silent", false, "Suppress all output except errors")
}
