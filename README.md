# Morphr

Fast RAW image converter for the command line. Convert CR2, NEF, ARW, DNG, and 25+ other camera RAW formats to JPEG, PNG, TIFF, WebP, or AVIF — one file or a whole folder.

Run `morphr` with no arguments to launch the interactive wizard.

---

## Installation

### Homebrew (recommended)

```bash
brew install idrewlong/tap/morphr
```

Installs morphr and all dependencies automatically. Then just run:

```bash
morphr
```

### From source

```bash
git clone https://github.com/idrewlong/morphr.git
cd morphr
make build
./morphr
```

Install dependencies manually:

```bash
# macOS
brew install dcraw exiftool webp libavif

# Ubuntu / Debian
sudo apt install dcraw libimage-exiftool-perl webp libavif-bin
```

---

## Interactive wizard

Run `morphr` with no arguments to launch the step-by-step wizard:

```
morphr
```

The wizard walks you through:

1. **Select RAW folder** — auto-detects nearby folders with RAW files, or browse with Finder
2. **Output format** — JPEG, PNG, TIFF, WebP, or AVIF
3. **Export folder** — defaults to `converted/` inside the input folder
4. **Include subfolders** — optionally recurse into subdirectories
5. **Quality** — High (95), Standard (85), or Web (75) — skipped for lossless formats
6. **Confirm** — summary of all settings, then "Let's go!" or Cancel

After conversion, each file is listed with its output size, total size and elapsed time are shown, and you're offered the option to open the export folder in Finder.

---

## CLI usage

### Convert a single file

```bash
morphr convert photo.CR2 -o photo.jpg
morphr convert photo.ARW --format png --quality 95 --max-width 2400
morphr convert shot.NEF -f webp -q 80
```

### Batch convert a directory

```bash
morphr batch ./raw-photos --format jpeg --quality 85 --output ./exports
morphr batch ./raw-photos -f webp -o ./exports --workers 8
morphr batch ./raw-photos --recursive --dry-run
```

### List supported formats

```bash
morphr formats
```

### Show RAW file metadata

```bash
morphr info photo.NEF
```

---

## Flags

### Global

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--format` | `-f` | `jpeg` | Output format: `jpeg` `png` `tiff` `webp` `avif` |
| `--quality` | `-q` | `90` | Quality 1–100 (JPEG, WebP, AVIF) |
| `--output` | `-o` | — | Output file or directory |
| `--overwrite` | | `false` | Overwrite existing files |
| `--verbose` | `-v` | `false` | Verbose logging |
| `--silent` | | `false` | Suppress all output except errors |

### convert + batch

| Flag | Default | Description |
|------|---------|-------------|
| `--max-width` | — | Fit longest edge to this pixel width |
| `--max-height` | — | Fit longest edge to this pixel height |
| `--scale` | — | Scale by percentage (e.g. `50`) |
| `--preserve-exif` | `true` | Copy EXIF/IPTC/XMP to output |
| `--auto-rotate` | `true` | Apply EXIF orientation tag |
| `--color-space` | `srgb` | Output color space: `srgb` `adobergb` `prophoto` |

### batch only

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--workers` | `-w` | NumCPU | Parallel workers |
| `--recursive` | `-r` | `false` | Recurse into subdirectories |
| `--dry-run` | | `false` | List files without converting |
| `--naming` | | `{name}` | Output naming template: `{name}` `{format}` |

---

## Supported formats

### Input (RAW)

| Camera brand | Extensions |
|---|---|
| Canon | `.cr2` `.cr3` `.crw` |
| Nikon | `.nef` `.nrw` |
| Sony | `.arw` `.sr2` `.srf` |
| Fujifilm | `.raf` |
| Olympus / OM System | `.orf` |
| Panasonic | `.rw2` |
| Pentax | `.pef` `.ptx` |
| Samsung | `.srw` |
| Adobe / Universal | `.dng` |
| Hasselblad | `.3fr` `.fff` |
| Phase One | `.iiq` |
| Leica | `.rwl` |
| Mamiya | `.mef` `.mos` |
| Minolta / Konica | `.mrw` |
| Kodak | `.dcr` `.k25` `.kdc` |
| RED | `.r3d` |
| ARRI | `.ari` |
| Generic | `.raw` `.bay` `.erf` |

### Output

| Format | Extension | Notes |
|---|---|---|
| JPEG | `.jpg` | Quality 1–100, universal |
| PNG | `.png` | Lossless |
| TIFF | `.tif` | Lossless, professional |
| WebP | `.webp` | Requires `cwebp` |
| AVIF | `.avif` | Requires `avifenc` |

---

## Architecture

```
morphr/
├── cmd/
│   ├── root.go          # Entry point — launches wizard with no args
│   ├── wizard.go        # Interactive wizard (huh + lipgloss)
│   ├── convert.go       # Single-file conversion
│   ├── batch.go         # Batch/directory conversion
│   ├── formats.go       # List supported formats
│   └── info.go          # Show RAW metadata
├── internal/
│   ├── config/          # Unified config struct
│   ├── decoder/
│   │   ├── decoder.go   # Decoder routing
│   │   ├── dcraw.go     # dcraw / dcraw_emu (LibRaw) subprocess
│   │   └── dng.go       # Pure Go DNG/TIFF decoder + EXIF extraction
│   ├── encoder/
│   │   ├── jpeg.go      # Pure Go
│   │   ├── png.go       # Pure Go
│   │   ├── tiff.go      # Pure Go
│   │   ├── webp.go      # cwebp subprocess
│   │   └── avif.go      # avifenc subprocess
│   ├── processor/
│   │   ├── resize.go    # Resize / fit-to-box
│   │   ├── rotate.go    # EXIF auto-rotation
│   │   └── color.go     # Color space conversion
│   ├── meta/            # EXIF/XMP read + preserve
│   ├── pipeline/
│   │   ├── walker.go    # File discovery
│   │   ├── pool.go      # Concurrent worker pool
│   │   └── progress.go  # Terminal progress bar
│   └── ui/
│       ├── logo.go      # MORPHR ASCII art + animated reveal
│       └── styles.go    # lipgloss color palette
├── .goreleaser.yaml
├── .github/workflows/release.yml
└── Makefile
```

---

## Development

```bash
make build          # Build binary → ./morphr
make test           # Run all tests
make bench          # Run benchmarks
make lint           # Run linter (golangci-lint)
make fmt            # Format code (gofmt + goimports)
make release-local  # Dry-run GoReleaser locally (no publish)
```

### Releasing

```bash
make tag-release V=1.0.0
```

This tags the commit, pushes to GitHub, and triggers the release workflow: tests → GoReleaser → GitHub Releases + Homebrew formula update.

---

## License

MIT — see [LICENSE](LICENSE).
