package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   lipgloss.Color
	ColorSecondary lipgloss.Color
	ColorSuccess   lipgloss.Color
	ColorWarning   lipgloss.Color
	ColorDanger    lipgloss.Color
	ColorMuted     lipgloss.Color
	ColorText      lipgloss.Color
	ColorSubtle    lipgloss.Color
	ColorDimBorder lipgloss.Color

	ActiveBorder       lipgloss.Style
	InactiveBorder     lipgloss.Style
	PanelTitle         lipgloss.Style
	InactivePanelTitle lipgloss.Style
	SelectedItem       lipgloss.Style
	NormalItem         lipgloss.Style
	MutedItem          lipgloss.Style
	WorkingBadge       lipgloss.Style
	DoneBadge          lipgloss.Style
	DeadBadge          lipgloss.Style
	TimestampStyle     lipgloss.Style
	StatusBar          lipgloss.Style
	HelpKey            lipgloss.Style
	HelpDesc           lipgloss.Style
	HelpSep            lipgloss.Style
	OverlayStyle       lipgloss.Style
	OverlayTitle       lipgloss.Style
	InputStyle         lipgloss.Style
	ErrorStyle         lipgloss.Style
	PreviewKey         lipgloss.Style
	PreviewValue       lipgloss.Style
	PreviewComment     lipgloss.Style
	CardIdle           lipgloss.Style
	CardWorking        lipgloss.Style
	CardDone           lipgloss.Style
	CardStopped        lipgloss.Style
	CardSelected       lipgloss.Style

	// WorkingPalette cycles through a gradient in sync with the spinner.
	WorkingPalette []lipgloss.Color
)

func init() {
	ApplyTheme(false)
}

// ApplyTheme rebuilds all UI styles for dark (light=false) or light (light=true) mode.
func ApplyTheme(light bool) {
	if light {
		ColorPrimary   = lipgloss.Color("#5B21B6")
		ColorSecondary = lipgloss.Color("#0369A1")
		ColorSuccess   = lipgloss.Color("#065F46")
		ColorWarning   = lipgloss.Color("#92400E")
		ColorDanger    = lipgloss.Color("#991B1B")
		ColorMuted     = lipgloss.Color("#6B7280")
		ColorText      = lipgloss.Color("#111827")
		ColorSubtle    = lipgloss.Color("#4B5563")
		ColorDimBorder = lipgloss.Color("#D1D5DB")
		WorkingPalette = []lipgloss.Color{
			"#6D28D9", "#4338CA", "#1D4ED8", "#0369A1",
			"#0891B2", "#0369A1", "#1D4ED8", "#4338CA",
		}
	} else {
		ColorPrimary   = lipgloss.Color("#7C3AED")
		ColorSecondary = lipgloss.Color("#06B6D4")
		ColorSuccess   = lipgloss.Color("#10B981")
		ColorWarning   = lipgloss.Color("#F59E0B")
		ColorDanger    = lipgloss.Color("#EF4444")
		ColorMuted     = lipgloss.Color("#6B7280")
		ColorText      = lipgloss.Color("#F9FAFB")
		ColorSubtle    = lipgloss.Color("#9CA3AF")
		ColorDimBorder = lipgloss.Color("#2D2D4E")
		WorkingPalette = []lipgloss.Color{
			"#BB9AF7", "#9D7CD8", "#7AA2F7", "#41A6B5",
			"#06B6D4", "#41A6B5", "#7AA2F7", "#9D7CD8",
		}
	}

	statusBarBg := lipgloss.Color("#1F2937")
	if light {
		statusBarBg = lipgloss.Color("#E5E7EB")
	}

	ActiveBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary)

	InactiveBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted)

	PanelTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary)

	InactivePanelTitle = lipgloss.NewStyle().
		Foreground(ColorMuted)

	SelectedItem = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	NormalItem = lipgloss.NewStyle().
		Foreground(ColorText)

	MutedItem = lipgloss.NewStyle().
		Foreground(ColorMuted)

	WorkingBadge = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true)

	DoneBadge = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	DeadBadge = lipgloss.NewStyle().
		Foreground(ColorDanger).
		Bold(true)

	TimestampStyle = lipgloss.NewStyle().
		Foreground(ColorSubtle)

	StatusBar = lipgloss.NewStyle().
		Foreground(ColorSubtle).
		Background(statusBarBg).
		Padding(0, 1)

	HelpKey = lipgloss.NewStyle().
		Foreground(ColorSecondary)

	HelpDesc = lipgloss.NewStyle().
		Foreground(ColorMuted)

	HelpSep = lipgloss.NewStyle().
		Foreground(ColorMuted)

	OverlayStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorWarning).
		Padding(1, 3)

	OverlayTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorWarning)

	InputStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ColorDanger).
		Bold(true)

	PreviewKey = lipgloss.NewStyle().
		Foreground(ColorSecondary).
		Bold(true)

	PreviewValue = lipgloss.NewStyle().
		Foreground(ColorText)

	PreviewComment = lipgloss.NewStyle().
		Foreground(ColorSubtle)

	CardIdle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Padding(0, 1)

	CardWorking = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorWarning).
		Padding(0, 1)

	CardDone = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSuccess).
		Padding(0, 1)

	CardStopped = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDimBorder).
		Padding(0, 1)

	CardSelected = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 1)
}
