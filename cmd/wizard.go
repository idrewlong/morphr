package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"

	"morphr/internal/config"
	"morphr/internal/pipeline"
	"morphr/internal/ui"
)

const browsePlaceholder = "__browse__"

// wizardKeyMap returns a huh KeyMap with both ctrl+c and esc bound to quit.
func wizardKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	return km
}

// stepResult controls the state machine flow.
type stepResult int

const (
	stepNext  stepResult = iota
	stepBack
	stepAbort
)

// wizardState holds all collected values across wizard steps.
// All folder paths are always stored as absolute paths.
type wizardState struct {
	inputFolder  string
	outputFolder string
	format       string
	quality      int
	recursive    bool
}

func runWizard() error {
	ui.PrintLogo()

	state := &wizardState{
		format:  "jpeg",
		quality: 85,
	}

	// Step indices:
	// 0: folder  1: format  2: output  3: recursive  4: quality  5: confirm
	steps := []func(*wizardState) stepResult{
		runStepFolder,
		runStepFormat,
		runStepOutput,
		runStepRecursive,
		runStepQuality,
		runStepConfirm,
	}

	i := 0
	for i < len(steps) {
		result := steps[i](state)
		switch result {
		case stepNext:
			i++
		case stepBack:
			if i == 0 {
				fmt.Println(ui.DimStyle.Render("\n  Cancelled."))
				os.Exit(130)
			}
			i--
			// Skip quality step when going backwards if format doesn't use quality
			if i == 4 && !formatUsesQuality(state.format) {
				i--
			}
		case stepAbort:
			fmt.Println(ui.DimStyle.Render("\n  Cancelled."))
			os.Exit(130)
		}
	}

	fmt.Println()
	return executeWizard(state)
}

// ── Step 1: Input folder ─────────────────────────────────────────────────────

func runStepFolder(state *wizardState) stepResult {
	detected := detectRAWFolders()

	for {
		if len(detected) == 0 {
			var choice string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Select RAW folder").
						Description("No RAW folders detected nearby.  Press Esc or Ctrl+C to cancel.").
						Options(
							huh.NewOption("Browse for folder...", "browse"),
							huh.NewOption("Enter folder path manually...", "manual"),
						).
						Value(&choice),
				),
			).WithTheme(huh.ThemeCharm()).WithKeyMap(wizardKeyMap())

			err := form.Run()
			if err == huh.ErrUserAborted {
				return stepBack
			}
			if err != nil {
				return stepAbort
			}

			if choice == "browse" {
				folder, cancelled := pickFolderFinder("Select your RAW folder")
				if cancelled {
					continue
				}
				if folder != "" {
					state.inputFolder = absPath(folder)
					return stepNext
				}
			} else {
				folder, cancelled := pickFolderManual()
				if cancelled {
					continue
				}
				if folder != "" {
					state.inputFolder = absPath(folder)
					return stepNext
				}
			}
			continue
		}

		var options []huh.Option[string]
		for _, f := range detected {
			label := fmt.Sprintf("%s  (%d RAW files)", f.path, f.count)
			options = append(options, huh.NewOption(label, f.path))
		}
		options = append(options, huh.NewOption("Browse for another folder...", browsePlaceholder))

		var choice string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select RAW folder").
					Description("Detected nearby.  Press Esc or Ctrl+C to cancel.").
					Options(options...).
					Value(&choice),
			),
		).WithTheme(huh.ThemeCharm()).WithKeyMap(wizardKeyMap())

		err := form.Run()
		if err == huh.ErrUserAborted {
			return stepBack
		}
		if err != nil {
			return stepAbort
		}

		if choice == browsePlaceholder {
			folder, cancelled := pickFolder("Select your RAW folder")
			if cancelled {
				continue
			}
			if folder != "" {
				state.inputFolder = absPath(folder)
				return stepNext
			}
			continue
		}

		state.inputFolder = absPath(choice)
		return stepNext
	}
}

// ── Step 2: Output format ────────────────────────────────────────────────────

func runStepFormat(state *wizardState) stepResult {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Output format").
				Description("Choose the format to convert your RAW files to.  Press Esc or Ctrl+C to go back.").
				Options(
					huh.NewOption("JPEG  (recommended, universal compatibility)", "jpeg"),
					huh.NewOption("PNG   (lossless)", "png"),
					huh.NewOption("TIFF  (professional, large files)", "tiff"),
					huh.NewOption("WebP  (modern web, great quality/size ratio)", "webp"),
					huh.NewOption("AVIF  (smallest files, slow — not for large batches)", "avif"),
				).
				Value(&state.format),
		),
	).WithTheme(huh.ThemeCharm()).WithKeyMap(wizardKeyMap())

	err := form.Run()
	if err == huh.ErrUserAborted {
		return stepBack
	}
	if err != nil {
		return stepAbort
	}
	return stepNext
}

// ── Step 3: Output folder ────────────────────────────────────────────────────

func runStepOutput(state *wizardState) stepResult {
	// inputFolder is already absolute so this default is always absolute too.
	defaultOutput := filepath.Join(state.inputFolder, "converted")

	for {
		var choice string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Export folder").
					Description("Where should converted files be saved?  Press Esc or Ctrl+C to go back.").
					Options(
						huh.NewOption(fmt.Sprintf("converted/  (created inside %s)", state.inputFolder), defaultOutput),
						huh.NewOption("Browse for a custom export location...", browsePlaceholder),
					).
					Value(&choice),
			),
		).WithTheme(huh.ThemeCharm()).WithKeyMap(wizardKeyMap())

		err := form.Run()
		if err == huh.ErrUserAborted {
			return stepBack
		}
		if err != nil {
			return stepAbort
		}

		if choice != browsePlaceholder {
			// choice is already absolute (it came from defaultOutput or a browse result)
			state.outputFolder = choice
			return stepNext
		}

		parentDir, cancelled := pickFolder("Select where to save converted files")
		if cancelled {
			continue
		}

		var folderName string
		nameForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Export folder name").
					Description(fmt.Sprintf("Will be created inside: %s", parentDir)).
					Placeholder("converted").
					Value(&folderName).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("name is required")
						}
						return nil
					}),
			),
		).WithTheme(huh.ThemeCharm()).WithKeyMap(wizardKeyMap())

		if err := nameForm.Run(); err == huh.ErrUserAborted {
			continue
		}

		// parentDir from pickFolder is already absolute
		state.outputFolder = filepath.Join(absPath(parentDir), folderName)
		return stepNext
	}
}

// ── Step 4: Recursive subfolders ─────────────────────────────────────────────

func runStepRecursive(state *wizardState) stepResult {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Include subfolders?").
				Description("Also convert RAW files in subdirectories.  Press Esc or Ctrl+C to go back.").
				Affirmative("Yes").
				Negative("No").
				Value(&state.recursive),
		),
	).WithTheme(huh.ThemeCharm()).WithKeyMap(wizardKeyMap())

	err := form.Run()
	if err == huh.ErrUserAborted {
		return stepBack
	}
	if err != nil {
		return stepAbort
	}
	return stepNext
}

// ── Step 5: Quality (skipped for lossless formats) ──────────────────────────

func runStepQuality(state *wizardState) stepResult {
	if !formatUsesQuality(state.format) {
		return stepNext
	}

	var qualityChoice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Quality").
				Description("Higher quality means larger files.  Press Esc or Ctrl+C to go back.").
				Options(
					huh.NewOption("High  (95) — best detail, largest files", "95"),
					huh.NewOption("Standard  (85) — balanced quality and size", "85"),
					huh.NewOption("Web  (75) — smaller files, slight compression", "75"),
				).
				Value(&qualityChoice),
		),
	).WithTheme(huh.ThemeCharm()).WithKeyMap(wizardKeyMap())

	err := form.Run()
	if err == huh.ErrUserAborted {
		return stepBack
	}
	if err != nil {
		return stepAbort
	}

	switch qualityChoice {
	case "95":
		state.quality = 95
	case "75":
		state.quality = 75
	default:
		state.quality = 85
	}
	return stepNext
}

// ── Step 5: Confirm ──────────────────────────────────────────────────────────

func runStepConfirm(state *wizardState) stepResult {
	files, _ := pipeline.WalkDirectory(state.inputFolder, false)
	printWizardSummary(state, len(files))

	confirmed := true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Start conversion?").
				Affirmative("Let's go!").
				Negative("Cancel").
				Value(&confirmed),
		),
	).WithTheme(huh.ThemeCharm()).WithKeyMap(wizardKeyMap())

	err := form.Run()
	if err == huh.ErrUserAborted {
		return stepBack
	}
	if err != nil {
		return stepAbort
	}
	if !confirmed {
		return stepBack
	}
	return stepNext
}

// ── Execute ──────────────────────────────────────────────────────────────────

func executeWizard(state *wizardState) error {
	cfg := config.Default()
	cfg.Format = state.format
	cfg.Quality = state.quality
	cfg.OutputPath = state.outputFolder
	cfg.Overwrite = true // user already confirmed via the wizard
	cfg.Workers = runtime.NumCPU()
	cfg.Silent = false

	files, err := pipeline.WalkDirectory(state.inputFolder, state.recursive)
	if err != nil {
		return fmt.Errorf("scan directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Println(ui.WarningStyle.Render("  No RAW files found."))
		return nil
	}

	fmt.Fprintf(os.Stderr, "  %s  %s\n\n",
		ui.LabelStyle.Render("Converting"),
		ui.TitleStyle.Render(fmt.Sprintf("%d RAW file(s)", len(files))),
	)

	if err := os.MkdirAll(state.outputFolder, 0o755); err != nil {
		return fmt.Errorf("create export folder: %w", err)
	}

	progress := pipeline.NewProgressReporter(len(files), false)
	ctx := context.Background()
	start := time.Now()

	type fileResult struct {
		entry pipeline.FileEntry
		err   error
	}
	results := make([]fileResult, 0, len(files))

	convertFn := func(ctx context.Context, entry pipeline.FileEntry, c *config.Config) error {
		fileCfg := *c
		fileCfg.InputPath = entry.Path
		fileCfg.Silent = true
		return convertFile(ctx, &fileCfg)
	}

	err = pipeline.RunPool(ctx, files, cfg, cfg.Workers, convertFn, func(r pipeline.Result) {
		progress.Update(r)
		results = append(results, fileResult{entry: r.Entry, err: r.Err})
	})
	progress.Finish()

	elapsed := time.Since(start).Round(time.Millisecond)

	// Print per-file summary with output sizes
	fmt.Println()
	var totalBytes int64
	var succeeded int
	for _, r := range results {
		inputName := filepath.Base(r.entry.Path)
		if r.err != nil {
			fmt.Printf("  %s  %s  %s\n",
				ui.ErrorStyle.Render("✗"),
				ui.FileStyle.Render(inputName),
				ui.DimStyle.Render(r.err.Error()),
			)
			continue
		}
		outPath := resolveOutputPath(r.entry.Path, cfg.OutputPath, cfg)
		sizeStr := "unknown"
		if info, statErr := os.Stat(outPath); statErr == nil {
			totalBytes += info.Size()
			sizeStr = formatSize(info.Size())
		}
		succeeded++
		fmt.Printf("  %s  %s  %s  %s\n",
			ui.SuccessStyle.Render("✓"),
			ui.FileStyle.Render(inputName),
			ui.DimStyle.Render("→"),
			ui.ValueStyle.Render(sizeStr),
		)
	}

	// Totals line
	fmt.Println()
	fmt.Printf("  %s  %s\n",
		ui.LabelStyle.Render("Total:"),
		ui.ValueStyle.Render(fmt.Sprintf("%d file(s) · %s · %s", succeeded, formatSize(totalBytes), elapsed)),
	)
	fmt.Println()

	// Offer to open the export folder in Finder (macOS only)
	if runtime.GOOS == "darwin" && succeeded > 0 {
		openIt := false
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Open export folder in Finder?").
					Affirmative("Open").
					Negative("Done").
					Value(&openIt),
			),
		).WithTheme(huh.ThemeCharm()).WithKeyMap(wizardKeyMap())
		if form.Run() == nil && openIt {
			exec.Command("open", state.outputFolder).Run()
		}
	}

	return err
}

// ── Summary ──────────────────────────────────────────────────────────────────

func printWizardSummary(state *wizardState, fileCount int) {
	qualityStr := "lossless"
	if formatUsesQuality(state.format) {
		qualityStr = fmt.Sprintf("%d", state.quality)
	}

	recursiveStr := "no"
	if state.recursive {
		recursiveStr = "yes (including subfolders)"
	}

	content := fmt.Sprintf(
		"  %s  %s\n"+
			"  %s  %s\n"+
			"  %s  %s\n"+
			"  %s  %s\n"+
			"  %s  %s\n"+
			"  %s  %s",
		ui.LabelStyle.Render("Input:"),
		ui.ValueStyle.Render(state.inputFolder),
		ui.LabelStyle.Render("Files:"),
		ui.TitleStyle.Render(fmt.Sprintf("%d RAW files found", fileCount)),
		ui.LabelStyle.Render("Format:"),
		ui.AccentStyle.Render(strings.ToUpper(state.format)),
		ui.LabelStyle.Render("Quality:"),
		ui.ValueStyle.Render(qualityStr),
		ui.LabelStyle.Render("Subfolders:"),
		ui.ValueStyle.Render(recursiveStr),
		ui.LabelStyle.Render("Export:"),
		ui.ValueStyle.Render(state.outputFolder),
	)

	fmt.Println(ui.HeaderBox.Render(content))
	fmt.Println()
}

// ── Folder picker ────────────────────────────────────────────────────────────

func pickFolder(prompt string) (string, bool) {
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("osascript"); err == nil {
			return pickFolderFinder(prompt)
		}
	}
	return pickFolderManual()
}

func pickFolderFinder(prompt string) (string, bool) {
	folder, err := macFolderDialog(prompt)
	if err != nil {
		return "", true
	}
	return folder, false
}

func pickFolderManual() (string, bool) {
	var folder string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Folder path").
				Description("Tip: paste a full path with Cmd+V, or drag a folder from Finder into this window.").
				Placeholder("/Users/you/RAW").
				Value(&folder).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("path is required")
					}
					info, err := os.Stat(s)
					if err != nil {
						return fmt.Errorf("cannot access: %s", s)
					}
					if !info.IsDir() {
						return fmt.Errorf("not a directory: %s", s)
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCharm()).WithKeyMap(wizardKeyMap())

	if err := form.Run(); err != nil {
		return "", true
	}
	return folder, false
}

func macFolderDialog(prompt string) (string, error) {
	script := fmt.Sprintf(`tell application "Finder" to activate
set chosenFolder to choose folder with prompt "%s"
return POSIX path of chosenFolder`, prompt)

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", err
	}

	folder := strings.TrimSpace(string(out))
	folder = strings.TrimRight(folder, "/")
	return folder, nil
}

// ── RAW folder detection ─────────────────────────────────────────────────────

type rawFolder struct {
	path  string
	count int
}

func detectRAWFolders() []rawFolder {
	var folders []rawFolder

	if count := countRAWFiles("."); count > 0 {
		folders = append(folders, rawFolder{path: ".", count: count})
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		return folders
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name()[0] == '.' {
			continue
		}
		path := entry.Name()
		if count := countRAWFiles(path); count > 0 {
			folders = append(folders, rawFolder{path: path, count: count})
		}
	}

	parentEntries, err := os.ReadDir("..")
	if err != nil {
		return folders
	}
	cwd, _ := os.Getwd()
	currentBase := filepath.Base(cwd)
	for _, entry := range parentEntries {
		if !entry.IsDir() || entry.Name()[0] == '.' || entry.Name() == currentBase {
			continue
		}
		path := filepath.Join("..", entry.Name())
		if count := countRAWFiles(path); count > 0 {
			folders = append(folders, rawFolder{path: path, count: count})
		}
	}

	return folders
}

func countRAWFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if config.SupportedInputFormats[ext] {
				count++
			}
		}
	}
	return count
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// absPath converts p to an absolute path. Falls back to p on error.
func absPath(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return abs
}

func formatUsesQuality(format string) bool {
	switch format {
	case "jpeg", "jpg", "webp", "avif":
		return true
	}
	return false
}
