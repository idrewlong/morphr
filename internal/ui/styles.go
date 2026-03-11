package ui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	ColorCyan    = lipgloss.Color("#00D4FF")
	ColorGreen   = lipgloss.Color("#00FF88")
	ColorYellow  = lipgloss.Color("#FFD93D")
	ColorRed     = lipgloss.Color("#FF6B6B")
	ColorGray    = lipgloss.Color("#888888")
	ColorWhite   = lipgloss.Color("#EEEEEE")
	ColorMagenta = lipgloss.Color("#C792EA")
	ColorOrange  = lipgloss.Color("#FF9F43")

	// Logo gradient colors
	ColorLogo1 = lipgloss.Color("#FF6B6B")
	ColorLogo2 = lipgloss.Color("#FF9F43")
	ColorLogo3 = lipgloss.Color("#FFD93D")
	ColorLogo4 = lipgloss.Color("#00FF88")
	ColorLogo5 = lipgloss.Color("#00D4FF")
	ColorLogo6 = lipgloss.Color("#C792EA")
)

// Text styles
var (
	TitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(ColorCyan)
	SuccessStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorGreen)
	WarningStyle = lipgloss.NewStyle().Foreground(ColorYellow)
	ErrorStyle   = lipgloss.NewStyle().Bold(true).Foreground(ColorRed)
	DimStyle     = lipgloss.NewStyle().Foreground(ColorGray)
	FileStyle    = lipgloss.NewStyle().Foreground(ColorWhite).Bold(true)
	StatStyle    = lipgloss.NewStyle().Foreground(ColorMagenta)
	LabelStyle   = lipgloss.NewStyle().Foreground(ColorGray)
	ValueStyle   = lipgloss.NewStyle().Foreground(ColorWhite)
	AccentStyle  = lipgloss.NewStyle().Foreground(ColorOrange).Bold(true)
)

// Box styles
var (
	SummaryBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	HeaderBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorGray).
			Padding(0, 2)
)
