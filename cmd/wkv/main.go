// wkv is a fuzzy-searchable TUI viewer for wezterm keybindings.
//
// It runs "wezterm show-keys", parses the output, and displays it
// in a color-coded, filterable table powered by Bubble Tea.
//
// # Usage
//
//	wkv
//
// Requires wezterm to be installed and available in your PATH.
//
// # Keybindings
//
//	j / ↓          Move cursor down
//	k / ↑          Move cursor up
//	g / Home       Go to top
//	G / End        Go to bottom
//	Ctrl+d         Half page down
//	Ctrl+u         Half page up
//	/              Start search
//	Escape         Exit search / clear filter
//	Tab            Next section filter
//	Shift+Tab      Previous section filter
//	q / Ctrl+c     Quit
//
// # Install
//
// Using Homebrew:
//
//	brew install sorafujitani/tap/wkv
//
// Using Go:
//
//	go install github.com/sorafujitani/wez-kv/cmd/wkv@latest
package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sorafujitani/wez-kv/internal/parser"
	"github.com/sorafujitani/wez-kv/internal/tui"
)

func main() {
	output, err := exec.Command("wezterm", "show-keys").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to run 'wezterm show-keys': %v\n", err)
		os.Exit(1)
	}

	result := parser.Parse(string(output))

	p := tea.NewProgram(tui.New(result), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
