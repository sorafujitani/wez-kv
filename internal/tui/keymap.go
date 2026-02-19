package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Top        key.Binding
	Bottom     key.Binding
	Search     key.Binding
	Escape     key.Binding
	NextTab    key.Binding
	PrevTab    key.Binding
	Quit       key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
	),
	Top: key.NewBinding(
		key.WithKeys("g", "home"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G", "end"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
	),
	NextTab: key.NewBinding(
		key.WithKeys("tab"),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("shift+tab"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
	),
	HalfPageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
	),
	HalfPageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
	),
}

type helpItem struct {
	key  string
	desc string
}

func helpItems() []helpItem {
	return []helpItem{
		{"j/k", "navigate"},
		{"/", "search"},
		{"Tab", "filter"},
		{"q", "quit"},
	}
}
