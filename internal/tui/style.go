package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("69"))

	leaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	leaderValueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("213"))

	tabBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("69")).
			Underline(true)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("252"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	selectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("236"))

	tableStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	actionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	searchPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("69")).
				Bold(true)

	matchCountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("69"))

	fuzzyMatchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("69")).
			Bold(true)
)

func modifierStyle(mod string) lipgloss.Style {
	switch mod {
	case "CTRL":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // Cyan
	case "SHIFT":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
	case "ALT":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // Magenta
	case "SUPER":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	}
}
