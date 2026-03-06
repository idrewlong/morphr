package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"morphr/internal/config"
	"morphr/internal/pipeline"
)

var batchCmd = &cobra.Command{
	Use:   "batch [directory]",
	Short: "Batch convert RAW files in a directory",
	Long: `Convert all RAW image files in a directory to the specified output format
with configurable concurrency.

Examples:
  morphr batch ./raw-photos --format jpeg --quality 85 --output ./exports
  morphr batch ./raw-photos -f png -o ./exports --workers 8
  morphr batch ./raw-photos --recursive --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runBatch,
}

func init() {
	batchCmd.Flags().IntP("workers", "w", 0, "Parallel workers (default: NumCPU)")
	batchCmd.Flags().BoolP("recursive", "r", false, "Recurse into subdirectories")
	batchCmd.Flags().Bool("dry-run", false, "List files that would be converted")
	batchCmd.Flags().String("naming", "{name}", "Output naming template: {name}, {format}")
	batchCmd.Flags().Int("max-width", 0, "Fit longest edge to pixel width")
	batchCmd.Flags().Int("max-height", 0, "Fit longest edge to pixel height")
	batchCmd.Flags().Float64("scale", 0, "Scale by percentage (e.g. 50)")
	batchCmd.Flags().Bool("preserve-exif", true, "Copy EXIF/IPTC/XMP to output")
	batchCmd.Flags().Bool("auto-rotate", true, "Apply EXIF orientation tag")
	batchCmd.Flags().String("color-space", "srgb", "Output color space: srgb|adobergb|prophoto")

	rootCmd.AddCommand(batchCmd)
}

func runBatch(cmd *cobra.Command, args []string) error {
	dir := args[0]

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("directory not found: %s", dir)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	cfg := buildConfig(cmd)

	if w, _ := cmd.Flags().GetInt("workers"); w > 0 {
		cfg.Workers = w
	}
	cfg.Recursive, _ = cmd.Flags().GetBool("recursive")
	cfg.DryRun, _ = cmd.Flags().GetBool("dry-run")
	if naming, _ := cmd.Flags().GetString("naming"); naming != "" {
		cfg.Naming = naming
	}

	files, err := pipeline.WalkDirectory(dir, cfg.Recursive)
	if err != nil {
		return fmt.Errorf("scan directory: %w", err)
	}

	if len(files) == 0 {
		if !cfg.Silent {
			fmt.Println("No RAW files found.")
		}
		return nil
	}

	if !cfg.Silent {
		fmt.Fprintf(os.Stderr, "Found %d RAW file(s)\n", len(files))
	}

	if cfg.DryRun {
		for _, f := range files {
			outPath := resolveOutputPath(f.Path, cfg.OutputPath, cfg)
			fmt.Printf("  %s → %s\n", f.Path, outPath)
		}
		return nil
	}

	progress := pipeline.NewProgressReporter(len(files), cfg.Silent)
	ctx := context.Background()

	convertFn := func(ctx context.Context, entry pipeline.FileEntry, c *config.Config) error {
		fileCfg := *c
		fileCfg.InputPath = entry.Path
		fileCfg.Silent = true
		return convertFile(ctx, &fileCfg)
	}

	err = pipeline.RunPool(ctx, files, cfg, cfg.Workers, convertFn, progress.Update)
	progress.Finish()

	return err
}
