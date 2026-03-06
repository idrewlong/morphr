package pipeline

import (
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
)

type ProgressReporter struct {
	bar    *progressbar.ProgressBar
	total  int
	done   int
	failed int
	silent bool
}

func NewProgressReporter(total int, silent bool) *ProgressReporter {
	if silent {
		return &ProgressReporter{total: total, silent: true}
	}

	bar := progressbar.NewOptions(total,
		progressbar.OptionSetDescription("Converting"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "│",
			BarEnd:        "│",
		}),
	)

	return &ProgressReporter{
		bar:   bar,
		total: total,
	}
}

func (p *ProgressReporter) Update(r Result) {
	if r.Err != nil {
		p.failed++
	}
	p.done++

	if !p.silent && p.bar != nil {
		p.bar.Add(1)
	}
}

func (p *ProgressReporter) Finish() {
	if !p.silent && p.bar != nil {
		p.bar.Finish()
	}
	if !p.silent {
		fmt.Fprintf(os.Stderr, "\nConverted %d/%d files", p.done-p.failed, p.total)
		if p.failed > 0 {
			fmt.Fprintf(os.Stderr, " (%d failed)", p.failed)
		}
		fmt.Fprintln(os.Stderr)
	}
}
