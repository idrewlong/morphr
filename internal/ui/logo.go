package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ASCII art logo lines for "MORPHR"
var logoLines = []string{
	`  ███╗   ███╗ ██████╗ ██████╗ ██████╗ ██╗  ██╗██████╗ `,
	`  ████╗ ████║██╔═══██╗██╔══██╗██╔══██╗██║  ██║██╔══██╗`,
	`  ██╔████╔██║██║   ██║██████╔╝██████╔╝███████║██████╔╝`,
	`  ██║╚██╔╝██║██║   ██║██╔══██╗██╔═══╝ ██╔══██║██╔══██╗`,
	`  ██║ ╚═╝ ██║╚██████╔╝██║  ██║██║     ██║  ██║██║  ██║`,
	`  ╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝`,
}

// Gradient colors for the logo reveal animation
var logoGradient = []lipgloss.Color{
	ColorLogo5, // blue
}

// PrintLogo prints the Morphr ASCII art logo with an animated reveal.
// Each line drops in with a gradient color, then a separator draws across.
func PrintLogo() {
	fmt.Println()

	for i, line := range logoLines {
		color := logoGradient[i%len(logoGradient)]
		style := lipgloss.NewStyle().Foreground(color).Bold(true)
		fmt.Println(style.Render(line))
		time.Sleep(45 * time.Millisecond)
	}

	sep := "─"
	width := 58
	for i := 1; i <= width; i++ {
		fmt.Printf("\r%s", DimStyle.Render(strings.Repeat(sep, i)))
		time.Sleep(4 * time.Millisecond)
	}
	fmt.Println()
	fmt.Println()
}
