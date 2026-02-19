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
