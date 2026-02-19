package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sorafujitani/wez-kv/internal/parser"
)

func testBindings() []parser.Keybinding {
	return []parser.Keybinding{
		{Table: "Default", Modifiers: "CTRL", Key: "c", Action: "CopyTo"},
		{Table: "Default", Modifiers: "CTRL", Key: "v", Action: "Paste"},
		{Table: "Default", Modifiers: "", Key: "Enter", Action: "ActivatePaneDirection"},
		{Table: "Copy", Modifiers: "CTRL", Key: "c", Action: "CopyMode"},
		{Table: "Copy", Modifiers: "", Key: "q", Action: "QuitCopy"},
		{Table: "Search", Modifiers: "", Key: "/", Action: "SearchForward"},
	}
}

func testResult() parser.ParseResult {
	return parser.ParseResult{
		Bindings: testBindings(),
		Tables:   []string{"Default", "Copy", "Search"},
		Leader:   &parser.Leader{Key: "a", Mods: "CTRL", Timeout: "1000ms"},
	}
}

func newTestModel() Model {
	m := New(testResult())
	// Set a reasonable terminal size
	m.width = 120
	m.height = 30
	return m
}

func sendKey(m Model, key string) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(Model)
}

func sendSpecialKey(m Model, keyType tea.KeyType) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: keyType})
	return updated.(Model)
}

func TestNew(t *testing.T) {
	m := newTestModel()

	if len(m.bindings) != 6 {
		t.Errorf("expected 6 bindings, got %d", len(m.bindings))
	}
	if len(m.filtered) != 6 {
		t.Errorf("expected 6 filtered, got %d", len(m.filtered))
	}
	if m.activeTable != -1 {
		t.Errorf("expected activeTable -1, got %d", m.activeTable)
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}
}

func TestCursorNavigation(t *testing.T) {
	m := newTestModel()

	// Move down
	m = sendKey(m, "j")
	if m.cursor != 1 {
		t.Errorf("after j: expected cursor 1, got %d", m.cursor)
	}

	// Move down again
	m = sendKey(m, "j")
	if m.cursor != 2 {
		t.Errorf("after jj: expected cursor 2, got %d", m.cursor)
	}

	// Move up
	m = sendKey(m, "k")
	if m.cursor != 1 {
		t.Errorf("after k: expected cursor 1, got %d", m.cursor)
	}

	// Go to top
	m = sendKey(m, "g")
	if m.cursor != 0 {
		t.Errorf("after g: expected cursor 0, got %d", m.cursor)
	}

	// Go to bottom
	m = sendKey(m, "G")
	if m.cursor != 5 {
		t.Errorf("after G: expected cursor 5, got %d", m.cursor)
	}
}

func TestCursorBounds(t *testing.T) {
	m := newTestModel()

	// Up at top should stay at 0
	m = sendKey(m, "k")
	if m.cursor != 0 {
		t.Errorf("up at top: expected cursor 0, got %d", m.cursor)
	}

	// Go to bottom then try down
	m = sendKey(m, "G")
	last := len(m.filtered) - 1
	m = sendKey(m, "j")
	if m.cursor != last {
		t.Errorf("down at bottom: expected cursor %d, got %d", last, m.cursor)
	}
}

func TestTabFilter(t *testing.T) {
	m := newTestModel()

	// All bindings initially
	if len(m.filtered) != 6 {
		t.Errorf("all: expected 6, got %d", len(m.filtered))
	}

	// Tab -> Default
	m = sendSpecialKey(m, tea.KeyTab)
	if m.activeTable != 0 {
		t.Errorf("expected activeTable 0, got %d", m.activeTable)
	}
	if len(m.filtered) != 3 {
		t.Errorf("Default: expected 3, got %d", len(m.filtered))
	}

	// Tab -> Copy
	m = sendSpecialKey(m, tea.KeyTab)
	if m.activeTable != 1 {
		t.Errorf("expected activeTable 1, got %d", m.activeTable)
	}
	if len(m.filtered) != 2 {
		t.Errorf("Copy: expected 2, got %d", len(m.filtered))
	}

	// Tab -> Search
	m = sendSpecialKey(m, tea.KeyTab)
	if m.activeTable != 2 {
		t.Errorf("expected activeTable 2, got %d", m.activeTable)
	}
	if len(m.filtered) != 1 {
		t.Errorf("Search: expected 1, got %d", len(m.filtered))
	}

	// Tab -> wraps back to All
	m = sendSpecialKey(m, tea.KeyTab)
	if m.activeTable != -1 {
		t.Errorf("expected activeTable -1, got %d", m.activeTable)
	}
	if len(m.filtered) != 6 {
		t.Errorf("All again: expected 6, got %d", len(m.filtered))
	}
}

func TestPrevTab(t *testing.T) {
	m := newTestModel()

	// Shift+Tab from All -> wraps to last table (Search)
	m = sendSpecialKey(m, tea.KeyShiftTab)
	if m.activeTable != 2 {
		t.Errorf("expected activeTable 2, got %d", m.activeTable)
	}
}

func TestSearchMode(t *testing.T) {
	m := newTestModel()

	// Enter search mode
	m = sendKey(m, "/")
	if !m.searching {
		t.Error("expected searching to be true")
	}

	// Exit search with Esc
	m = sendSpecialKey(m, tea.KeyEsc)
	if m.searching {
		t.Error("expected searching to be false after Esc")
	}

	// Enter search mode and confirm with Enter
	m = sendKey(m, "/")
	m = sendSpecialKey(m, tea.KeyEnter)
	if m.searching {
		t.Error("expected searching to be false after Enter")
	}
}

func TestApplyFilterWithQuery(t *testing.T) {
	m := newTestModel()

	// Simulate a search query
	m.query = "Paste"
	m.applyFilter()

	if len(m.filtered) != 1 {
		t.Errorf("expected 1 match for 'Paste', got %d", len(m.filtered))
	}
	if m.filtered[0].Action != "Paste" {
		t.Errorf("expected action 'Paste', got %q", m.filtered[0].Action)
	}
}

func TestApplyFilterTableAndQuery(t *testing.T) {
	m := newTestModel()

	// Filter by Copy table
	m.activeTable = 1
	m.query = "CopyMode"
	m.applyFilter()

	if len(m.filtered) != 1 {
		t.Errorf("expected 1 match, got %d", len(m.filtered))
	}
	if m.filtered[0].Table != "Copy" {
		t.Errorf("expected table 'Copy', got %q", m.filtered[0].Table)
	}
}

func TestEscapeClearsQueryThenTable(t *testing.T) {
	m := newTestModel()

	// Set a query and table filter
	m.activeTable = 0
	m.query = "test"
	m.searchInput.SetValue("test")
	m.applyFilter()

	// First Esc clears query
	m = sendSpecialKey(m, tea.KeyEsc)
	if m.query != "" {
		t.Errorf("expected query cleared, got %q", m.query)
	}
	if m.activeTable != 0 {
		t.Errorf("expected activeTable still 0, got %d", m.activeTable)
	}

	// Second Esc clears table filter
	m = sendSpecialKey(m, tea.KeyEsc)
	if m.activeTable != -1 {
		t.Errorf("expected activeTable -1, got %d", m.activeTable)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m := newTestModel()

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	m = updated.(Model)

	if m.width != 200 {
		t.Errorf("expected width 200, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height 50, got %d", m.height)
	}
}

func TestVisibleRows(t *testing.T) {
	m := newTestModel()

	m.height = 20
	expected := 20 - 8 // overhead = 8
	if v := m.visibleRows(); v != expected {
		t.Errorf("expected %d visible rows, got %d", expected, v)
	}

	// Very small height should return at least 1
	m.height = 5
	if v := m.visibleRows(); v != 1 {
		t.Errorf("expected 1 visible row for small height, got %d", v)
	}
}

func TestViewEmptyWidth(t *testing.T) {
	m := New(testResult())
	// width=0 should return empty
	if v := m.View(); v != "" {
		t.Errorf("expected empty view for width=0, got %q", v)
	}
}

func TestViewRenders(t *testing.T) {
	m := newTestModel()
	v := m.View()

	if v == "" {
		t.Error("expected non-empty view")
	}
}

func TestCursorResetOnFilter(t *testing.T) {
	m := newTestModel()

	// Move cursor down
	m = sendKey(m, "j")
	m = sendKey(m, "j")
	if m.cursor != 2 {
		t.Fatalf("expected cursor 2, got %d", m.cursor)
	}

	// Switch tab -> cursor resets
	m = sendSpecialKey(m, tea.KeyTab)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after tab switch, got %d", m.cursor)
	}
}
