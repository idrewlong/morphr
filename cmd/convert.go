package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"morphr/internal/config"
	"morphr/internal/decoder"
	"morphr/internal/encoder"
	"morphr/internal/meta"
	"morphr/internal/processor"
)

var convertCmd = &cobra.Command{
	Use:   "convert [file]",
	Short: "Convert a single RAW file",
	Long: `Convert a single RAW image file to the specified output format.

Examples:
  morphr convert photo.CR2 -o photo.jpg
  morphr convert photo.ARW --format png --quality 95 --max-width 2400
  morphr convert shot.NEF -f webp -q 80`,
	Args: cobra.ExactArgs(1),
	RunE: runConvert,
}

func init() {
	convertCmd.Flags().Int("max-width", 0, "Fit longest edge to pixel width")
	convertCmd.Flags().Int("max-height", 0, "Fit longest edge to pixel height")
	convertCmd.Flags().Float64("scale", 0, "Scale by percentage (e.g. 50)")
	convertCmd.Flags().Bool("preserve-exif", true, "Copy EXIF/IPTC/XMP to output")
	convertCmd.Flags().Bool("auto-rotate", true, "Apply EXIF orientation tag")
	convertCmd.Flags().String("color-space", "srgb", "Output color space: srgb|adobergb|prophoto")

	rootCmd.AddCommand(convertCmd)
}

func runConvert(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", inputPath)
	}

	ext := strings.ToLower(filepath.Ext(inputPath))
	if !config.SupportedInputFormats[ext] {
		return fmt.Errorf("unsupported input format: %s", ext)
	}

	cfg := buildConfig(cmd)
	cfg.InputPath = inputPath

	return convertFile(context.Background(), cfg)
}

func buildConfig(cmd *cobra.Command) *config.Config {
	cfg := config.Default()

	if f, _ := cmd.Flags().GetString("format"); f != "" {
		cfg.Format = strings.ToLower(f)
	}
	if q, _ := cmd.Flags().GetInt("quality"); q > 0 {
		cfg.Quality = q
	}
	if o, _ := cmd.Flags().GetString("output"); o != "" {
		cfg.OutputPath = o
	}
	cfg.Overwrite, _ = cmd.Flags().GetBool("overwrite")
	cfg.Verbose, _ = cmd.Flags().GetBool("verbose")
	cfg.Silent, _ = cmd.Flags().GetBool("silent")

	if cmd.Flags().Lookup("max-width") != nil {
		cfg.MaxWidth, _ = cmd.Flags().GetInt("max-width")
	}
	if cmd.Flags().Lookup("max-height") != nil {
		cfg.MaxHeight, _ = cmd.Flags().GetInt("max-height")
	}
	if cmd.Flags().Lookup("scale") != nil {
		cfg.Scale, _ = cmd.Flags().GetFloat64("scale")
	}
	if cmd.Flags().Lookup("preserve-exif") != nil {
		cfg.PreserveExif, _ = cmd.Flags().GetBool("preserve-exif")
	}
	if cmd.Flags().Lookup("auto-rotate") != nil {
		cfg.AutoRotate, _ = cmd.Flags().GetBool("auto-rotate")
	}
	if cmd.Flags().Lookup("color-space") != nil {
		cfg.ColorSpace, _ = cmd.Flags().GetString("color-space")
	}

	return cfg
}

func convertFile(ctx context.Context, cfg *config.Config) error {
	start := time.Now()
	verbose := cfg.Verbose && !cfg.Silent
	outputPath := resolveOutputPath(cfg.InputPath, cfg.OutputPath, cfg)

	if !cfg.Overwrite {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("output file already exists: %s (use --overwrite to replace)", outputPath)
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Decoding %s...\n", filepath.Base(cfg.InputPath))
	}

	result, err := decoder.Decode(cfg.InputPath)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	img := result.Image

	if cfg.AutoRotate && result.Meta != nil && result.Meta.Orientation > 1 {
		if verbose {
			fmt.Fprintf(os.Stderr, "Auto-rotating (orientation %d)...\n", result.Meta.Orientation)
		}
		img = processor.AutoRotate(img, result.Meta.Orientation)
	}

	if cfg.MaxWidth > 0 || cfg.MaxHeight > 0 || cfg.Scale > 0 {
		if verbose {
			fmt.Fprintln(os.Stderr, "Resizing...")
		}
		img = processor.Resize(img, cfg.MaxWidth, cfg.MaxHeight, cfg.Scale)
	}

	if cfg.ColorSpace != "" && cfg.ColorSpace != "srgb" {
		if verbose {
			fmt.Fprintf(os.Stderr, "Converting to %s...\n", cfg.ColorSpace)
		}
		img = processor.ConvertColorSpace(img, cfg.ColorSpace)
	}

	enc, err := encoder.Get(cfg.Format)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Encoding %s...\n", enc.Format())
	}

	if dir := filepath.Dir(outputPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer outFile.Close()

	opts := encoder.Options{Quality: cfg.Quality}
	if err := enc.Encode(outFile, img, opts); err != nil {
		outFile.Close()
		os.Remove(outputPath)
		return fmt.Errorf("encode: %w", err)
	}
	outFile.Close()

	if cfg.PreserveExif {
		if err := meta.PreserveEXIF(cfg.InputPath, outputPath); err != nil && verbose {
			fmt.Fprintf(os.Stderr, "Warning: EXIF preservation skipped: %v\n", err)
		}
	}

	if !cfg.Silent {
		elapsed := time.Since(start)
		size := "unknown"
		if info, err := os.Stat(outputPath); err == nil {
			size = formatSize(info.Size())
		}
		fmt.Printf("%s → %s (%s, %s)\n",
			filepath.Base(cfg.InputPath),
			filepath.Base(outputPath),
			size,
			elapsed.Round(time.Millisecond),
		)
	}

	return nil
}

func resolveOutputPath(inputPath, outputFlag string, cfg *config.Config) string {
	ext := cfg.OutputExtension()
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))

	outName := cfg.Naming
	outName = strings.ReplaceAll(outName, "{name}", baseName)
	outName = strings.ReplaceAll(outName, "{format}", strings.TrimPrefix(ext, "."))

	if outputFlag != "" {
		if filepath.Ext(outputFlag) != "" {
			return outputFlag
		}
		return filepath.Join(outputFlag, outName+ext)
	}

	return filepath.Join(filepath.Dir(inputPath), outName+ext)
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
