package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"github.com/sorafujitani/wez-kv/internal/parser"
)

type Model struct {
	bindings     []parser.Keybinding
	filtered     []parser.Keybinding
	matchIndices [][]int // fuzzy match indices per filtered row
	tables       []string
	leader       *parser.Leader

	cursor      int
	offset      int
	width       int
	height      int
	activeTable int // -1 = All
	searching   bool
	searchInput textinput.Model
	query       string
}

func New(result parser.ParseResult) Model {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.CharLimit = 128

	m := Model{
		bindings:    result.Bindings,
		tables:      result.Tables,
		leader:      result.Leader,
		activeTable: -1,
		searchInput: ti,
	}
	m.applyFilter()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.clampView()
		return m, nil

	case tea.KeyMsg:
		if m.searching {
			return m.updateSearch(msg)
		}
		return m.updateNormal(msg)
	}
	return m, nil
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, keys.Down):
		m.cursorDown()
	case key.Matches(msg, keys.Up):
		m.cursorUp()
	case key.Matches(msg, keys.Top):
		m.cursor = 0
		m.offset = 0
	case key.Matches(msg, keys.Bottom):
		m.cursor = max(0, len(m.filtered)-1)
		m.clampView()
	case key.Matches(msg, keys.HalfPageDown):
		half := m.visibleRows() / 2
		for range half {
			m.cursorDown()
		}
	case key.Matches(msg, keys.HalfPageUp):
		half := m.visibleRows() / 2
		for range half {
			m.cursorUp()
		}
	case key.Matches(msg, keys.Search):
		m.searching = true
		m.searchInput.Focus()
		return m, textinput.Blink
	case key.Matches(msg, keys.Escape):
		if m.query != "" {
			m.query = ""
			m.searchInput.SetValue("")
			m.applyFilter()
		} else if m.activeTable != -1 {
			m.activeTable = -1
			m.applyFilter()
		}
	case key.Matches(msg, keys.NextTab):
		m.activeTable++
		if m.activeTable >= len(m.tables) {
			m.activeTable = -1
		}
		m.applyFilter()
	case key.Matches(msg, keys.PrevTab):
		m.activeTable--
		if m.activeTable < -1 {
			m.activeTable = len(m.tables) - 1
		}
		m.applyFilter()
	}
	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searching = false
		m.searchInput.Blur()
		return m, nil
	case tea.KeyEnter:
		m.searching = false
		m.searchInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.query = m.searchInput.Value()
	m.applyFilter()
	return m, cmd
}

func (m *Model) applyFilter() {
	// First filter by table
	var candidates []parser.Keybinding
	var candidateIndices []int
	for i, b := range m.bindings {
		if m.activeTable == -1 || (m.activeTable < len(m.tables) && b.Table == m.tables[m.activeTable]) {
			candidates = append(candidates, b)
			candidateIndices = append(candidateIndices, i)
		}
	}

	m.matchIndices = nil

	if m.query == "" {
		m.filtered = candidates
		m.matchIndices = make([][]int, len(candidates))
	} else {
		// Build searchable strings
		strs := make([]string, len(candidates))
		for i, b := range candidates {
			strs[i] = b.Modifiers + " " + b.Key + " " + b.Action
		}

		matches := fuzzy.Find(m.query, strs)
		m.filtered = make([]parser.Keybinding, len(matches))
		m.matchIndices = make([][]int, len(matches))
		for i, match := range matches {
			m.filtered[i] = candidates[match.Index]
			m.matchIndices[i] = match.MatchedIndexes
		}
	}

	m.cursor = 0
	m.offset = 0
}

func (m *Model) cursorDown() {
	if m.cursor < len(m.filtered)-1 {
		m.cursor++
		m.clampView()
	}
}

func (m *Model) cursorUp() {
	if m.cursor > 0 {
		m.cursor--
		m.clampView()
	}
}

func (m *Model) clampView() {
	visible := m.visibleRows()
	if visible <= 0 {
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

func (m Model) visibleRows() int {
	// header(1) + tabBar(1) + separator(1) + columnHeader(1) + separator(1) + ... + separator(1) + search(1) + help(1)
	overhead := 8
	rows := m.height - overhead
	if rows < 1 {
		return 1
	}
	return rows
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Title + Leader
	b.WriteString(m.renderTitle())
	b.WriteString("\n")

	// Tab bar
	b.WriteString(m.renderTabBar())
	b.WriteString("\n")

	// Separator
	b.WriteString(m.renderSeparator())
	b.WriteString("\n")

	// Column header
	b.WriteString(m.renderColumnHeader())
	b.WriteString("\n")

	// Separator
	b.WriteString(m.renderSeparator())
	b.WriteString("\n")

	// Rows
	visible := m.visibleRows()
	end := min(m.offset+visible, len(m.filtered))
	for i := m.offset; i < end; i++ {
		row := m.renderRow(i)
		b.WriteString(row)
		b.WriteString("\n")
	}

	// Pad remaining lines
	for i := end - m.offset; i < visible; i++ {
		b.WriteString("\n")
	}

	// Separator
	b.WriteString(m.renderSeparator())
	b.WriteString("\n")

	// Search bar
	b.WriteString(m.renderSearchBar())
	b.WriteString("\n")

	// Help bar
	b.WriteString(m.renderHelp())

	return b.String()
}

func (m Model) renderTitle() string {
	title := titleStyle.Render(" wez-kv")
	if m.leader == nil {
		return title
	}

	var leaderParts []string
	if m.leader.Mods != "" {
		leaderParts = append(leaderParts, m.leader.Mods)
	}
	leaderParts = append(leaderParts, m.leader.Key)
	leaderStr := leaderValueStyle.Render(strings.Join(leaderParts, "+"))
	timeout := leaderStyle.Render(fmt.Sprintf("(%s)", m.leader.Timeout))

	right := leaderStyle.Render("Leader: ") + leaderStr + " " + timeout
	gap := m.width - lipgloss.Width(title) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return title + strings.Repeat(" ", gap) + right
}

func (m Model) renderTabBar() string {
	var parts []string

	if m.activeTable == -1 {
		parts = append(parts, activeTabStyle.Render(" [All]"))
	} else {
		parts = append(parts, tabBarStyle.Render(" [All]"))
	}

	for i, t := range m.tables {
		if i == m.activeTable {
			parts = append(parts, activeTabStyle.Render(t))
		} else {
			parts = append(parts, tabBarStyle.Render(t))
		}
	}

	return " " + strings.Join(parts, "  ")
}

func (m Model) renderSeparator() string {
	return separatorStyle.Render(" " + strings.Repeat("─", max(0, m.width-2)))
}

func (m Model) renderColumnHeader() string {
	return headerStyle.Render(m.formatColumns("Table", "Modifiers", "Key", "Action"))
}

func (m Model) colWidths() (int, int, int, int) {
	tW := 18
	mW := 18
	kW := 20
	aW := m.width - tW - mW - kW - 5 // 5 = leading space + 3 separators + trailing
	if aW < 20 {
		aW = 20
	}
	return tW, mW, kW, aW
}

func (m Model) formatColumns(table, mods, key, action string) string {
	tW, mW, kW, _ := m.colWidths()
	return fmt.Sprintf(" %-*s %-*s %-*s %s", tW, table, mW, mods, kW, key, action)
}

func (m Model) renderRow(idx int) string {
	b := m.filtered[idx]
	selected := idx == m.cursor

	table := tableStyle.Render(b.Table)
	mods := renderModifiers(b.Modifiers)
	k := keyStyle.Render(b.Key)
	action := actionStyle.Render(b.Action)

	tW, mW, kW, _ := m.colWidths()

	// Pad columns accounting for ANSI codes
	tablePad := tW - lipgloss.Width(table)
	modsPad := mW - lipgloss.Width(mods)
	keyPad := kW - lipgloss.Width(k)
	if tablePad < 0 {
		tablePad = 0
	}
	if modsPad < 0 {
		modsPad = 0
	}
	if keyPad < 0 {
		keyPad = 0
	}

	row := " " + table + strings.Repeat(" ", tablePad) + " " +
		mods + strings.Repeat(" ", modsPad) + " " +
		k + strings.Repeat(" ", keyPad) + " " +
		action

	if selected {
		// Apply background to the full width
		padLen := m.width - lipgloss.Width(row)
		if padLen > 0 {
			row += strings.Repeat(" ", padLen)
		}
		row = selectedRowStyle.Render(row)
	}

	return row
}

func renderModifiers(mods string) string {
	if mods == "" {
		return ""
	}
	parts := strings.Split(mods, " | ")
	var rendered []string
	for _, p := range parts {
		rendered = append(rendered, modifierStyle(p).Render(p))
	}
	return strings.Join(rendered, lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render(" | "))
}

func (m Model) renderSearchBar() string {
	if m.searching {
		input := m.searchInput.View()
		count := matchCountStyle.Render(fmt.Sprintf("%d/%d matches", len(m.filtered), len(m.bindings)))
		gap := m.width - lipgloss.Width(input) - lipgloss.Width(count) - 2
		if gap < 1 {
			gap = 1
		}
		return " " + input + strings.Repeat(" ", gap) + count
	}

	if m.query != "" {
		prompt := searchPromptStyle.Render("> ") + m.query
		count := matchCountStyle.Render(fmt.Sprintf("%d/%d matches", len(m.filtered), len(m.bindings)))
		gap := m.width - lipgloss.Width(prompt) - lipgloss.Width(count) - 2
		if gap < 1 {
			gap = 1
		}
		return " " + prompt + strings.Repeat(" ", gap) + count
	}

	count := matchCountStyle.Render(fmt.Sprintf("%d entries", len(m.filtered)))
	return " " + count
}

func (m Model) renderHelp() string {
	items := helpItems()
	var parts []string
	for _, item := range items {
		parts = append(parts, helpKeyStyle.Render(item.key)+helpStyle.Render(":"+item.desc))
	}
	return " " + strings.Join(parts, helpStyle.Render("  "))
}
