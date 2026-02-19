package parser

import (
	"regexp"
	"strings"
)

type Keybinding struct {
	Table     string
	Modifiers string
	Key       string
	Action    string
}

type Leader struct {
	Key     string
	Mods    string
	Timeout string
}

type ParseResult struct {
	Leader   *Leader
	Bindings []Keybinding
	Tables   []string
}

var (
	leaderRe  = regexp.MustCompile(`^Leader:\s+(.+?)\s+((?:CTRL|SHIFT|ALT|SUPER|NONE)(?:\s*\|\s*(?:CTRL|SHIFT|ALT|SUPER|NONE))*)\s+(\S+)$`)
	separatorRe = regexp.MustCompile(`\s+->\s+`)
	modsKeyRe   = regexp.MustCompile(`^((?:CTRL|SHIFT|ALT|SUPER|NONE)(?:\s*\|\s*(?:CTRL|SHIFT|ALT|SUPER|NONE))*)\s+(.+)$`)
)

func Parse(input string) ParseResult {
	var result ParseResult
	var currentTable string
	tablesSeen := make(map[string]bool)

	lines := strings.Split(input, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Leader:") {
			if m := leaderRe.FindStringSubmatch(line); m != nil {
				mods := m[2]
				if mods == "NONE" {
					mods = ""
				}
				result.Leader = &Leader{
					Key:     m[1],
					Mods:    mods,
					Timeout: m[3],
				}
			}
			continue
		}

		if line == "Default key table" {
			currentTable = "Default"
			if !tablesSeen[currentTable] {
				result.Tables = append(result.Tables, currentTable)
				tablesSeen[currentTable] = true
			}
			continue
		}
		if strings.HasPrefix(line, "Key Table: ") {
			currentTable = strings.TrimPrefix(line, "Key Table: ")
			if !tablesSeen[currentTable] {
				result.Tables = append(result.Tables, currentTable)
				tablesSeen[currentTable] = true
			}
			continue
		}
		if line == "Mouse" || strings.HasPrefix(line, "Mouse: ") {
			currentTable = line
			if !tablesSeen[currentTable] {
				result.Tables = append(result.Tables, currentTable)
				tablesSeen[currentTable] = true
			}
			continue
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "---") {
			continue
		}

		if !strings.HasPrefix(line, "\t") {
			continue
		}

		loc := separatorRe.FindStringIndex(line)
		if loc == nil {
			continue
		}

		left := strings.TrimSpace(line[:loc[0]])
		action := strings.TrimSpace(line[loc[1]:])

		var modifiers, key string
		if m := modsKeyRe.FindStringSubmatch(left); m != nil {
			modifiers = normalizeModifiers(m[1])
			key = strings.TrimSpace(m[2])
		} else {
			key = left
		}

		result.Bindings = append(result.Bindings, Keybinding{
			Table:     currentTable,
			Modifiers: modifiers,
			Key:       key,
			Action:    action,
		})
	}

	return result
}

func normalizeModifiers(raw string) string {
	parts := strings.Split(raw, "|")
	var mods []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" && p != "NONE" {
			mods = append(mods, p)
		}
	}
	return strings.Join(mods, " | ")
}
