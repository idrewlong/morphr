package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"morphr/internal/config"
)

var formatsCmd = &cobra.Command{
	Use:   "formats",
	Short: "List supported input and output formats",
	RunE:  runFormats,
}

func init() {
	rootCmd.AddCommand(formatsCmd)
}

func runFormats(cmd *cobra.Command, args []string) error {
	fmt.Println("Supported Input Formats (RAW):")
	fmt.Println(strings.Repeat("─", 50))

	var inputs []string
	for ext := range config.SupportedInputFormats {
		inputs = append(inputs, strings.ToUpper(strings.TrimPrefix(ext, ".")))
	}
	sort.Strings(inputs)

	for i, ext := range inputs {
		if i > 0 && i%10 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-6s", ext)
	}
	fmt.Println()

	fmt.Println()
	fmt.Println("Supported Output Formats:")
	fmt.Println(strings.Repeat("─", 50))

	outputs := []struct{ name, ext, note string }{
		{"JPEG", ".jpg", "lossy, quality 1-100"},
		{"PNG", ".png", "lossless"},
		{"TIFF", ".tif", "lossless, deflate compression"},
		{"WebP", ".webp", "lossy, requires cwebp"},
		{"AVIF", ".avif", "lossy, requires avifenc"},
	}

	for _, o := range outputs {
		fmt.Printf("  %-6s (%s)  %s\n", o.name, o.ext, o.note)
	}

	return nil
}
