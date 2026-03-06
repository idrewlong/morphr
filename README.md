# Morphr

A fast, dependency-light command-line tool for converting camera RAW image files into standard formats.

**Supported RAW formats:** CR2, CR3, NEF, ARW, DNG, RAF, ORF, RW2, PEF, SRW, and 20+ more.
**Output formats:** JPEG, PNG, TIFF, WebP, AVIF.

## Installation

### Homebrew (recommended)

```bash
brew install idrewlong/tap/morphr
```

This automatically installs all dependencies (dcraw, exiftool, webp, libavif). After install, run with:

```bash
morphr convert photo.CR2 -o photo.jpg
```

### From source

```bash
git clone https://github.com/idrewlong/morphr.git
cd morphr
make build
```

Then run locally with:

```bash
./morphr convert photo.CR2 -o photo.jpg
```

You'll need the dependencies installed separately:

```bash
# macOS
brew install dcraw exiftool webp libavif

# Ubuntu / Debian
sudo apt install dcraw libimage-exiftool-perl webp libavif-bin
```

### Go install

```bash
go install github.com/idrewlong/morphr@latest
```

## Usage

### Convert a single file

```bash
morphr convert photo.CR2 -o photo.jpg
morphr convert photo.ARW --format png --quality 95 --max-width 2400
morphr convert shot.NEF -f webp -q 80
```

### Batch convert a directory

```bash
morphr batch ./raw-photos --format jpeg --quality 85 --output ./exports
morphr batch ./raw-photos -f png -o ./exports --workers 8
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

## Flags

### Global

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--format` | `-f` | Output format: jpeg, png, tiff, webp, avif | jpeg |
| `--quality` | `-q` | JPEG/WebP quality 1-100 | 90 |
| `--output` | `-o` | Output file or directory | — |
| `--overwrite` | | Overwrite existing files | false |
| `--verbose` | `-v` | Verbose logging | false |
| `--silent` | | Suppress all output except errors | false |

### Convert / Batch

| Flag | Description | Default |
|------|-------------|---------|
| `--max-width` | Fit longest edge to pixel width | — |
| `--max-height` | Fit longest edge to pixel height | — |
| `--scale` | Scale by percentage (e.g. 50) | — |
| `--preserve-exif` | Copy EXIF/IPTC/XMP to output | true |
| `--auto-rotate` | Apply EXIF orientation tag | true |
| `--color-space` | Output color space: srgb, adobergb, prophoto | srgb |

### Batch only

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--workers` | `-w` | Parallel workers | NumCPU |
| `--recursive` | `-r` | Recurse into subdirectories | false |
| `--dry-run` | | List files without converting | false |
| `--naming` | | Output naming template: `{name}`, `{format}` | `{name}` |

## Architecture

```
morphr/
├── cmd/                    # CLI commands (cobra)
│   ├── root.go
│   ├── convert.go          # Single-file conversion
│   ├── batch.go            # Batch/directory conversion
│   ├── formats.go          # List supported formats
│   └── info.go             # Show RAW metadata
├── internal/
│   ├── config/             # Unified config struct
│   ├── decoder/            # RAW → image.Image
│   │   ├── dcraw.go        # dcraw/LibRaw subprocess
│   │   └── dng.go          # Pure Go DNG/TIFF decoder
│   ├── processor/          # Image transformations
│   │   ├── resize.go       # Resize / fit-to-box
│   │   ├── rotate.go       # EXIF auto-rotation
│   │   └── color.go        # Color space conversion
│   ├── encoder/            # image.Image → output format
│   │   ├── jpeg.go, png.go, tiff.go
│   │   ├── webp.go         # via cwebp subprocess
│   │   └── avif.go         # via avifenc subprocess
│   ├── meta/               # EXIF/XMP metadata
│   └── pipeline/           # Concurrent processing
│       ├── walker.go       # File discovery
│       ├── pool.go         # Worker pool
│       └── progress.go     # Terminal progress bar
├── Makefile
├── .goreleaser.yaml
└── README.md
```

## Development

```bash
make build          # Build binary → ./morphr
make test           # Run all tests
make bench          # Run benchmarks
make lint           # Run linter
make fmt            # Format code
make release-local  # Test GoReleaser config (no publish)
```

### Releasing

```bash
make tag-release V=0.1.0
```

This tags the commit and pushes to GitHub, which triggers the release workflow:
build → test → GoReleaser → GitHub Releases + Homebrew formula update.

## License

MIT — see [LICENSE](LICENSE).
