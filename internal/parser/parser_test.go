package parser

import (
	"testing"
)

const testInput = `Leader: Char('a') CTRL 2.001s
Default key table
-----------------

	CTRL                 Tab                ->   ActivateTabRelative(1)
	SHIFT | CTRL         Tab                ->   ActivateTabRelative(-1)
	SHIFT                Enter              ->   SendString("\n")
	ALT                  Enter              ->   ToggleFullScreen
	SHIFT | ALT | CTRL   DownArrow          ->   AdjustPaneSize(Down, 1)
	                     Copy               ->   CopyTo(Clipboard)
	                     Paste              ->   PasteFrom(Clipboard)

Key Table: copy_mode
--------------------

	        Tab          ->   CopyMode(MoveForwardWord)
	SHIFT   Tab          ->   CopyMode(MoveBackwardWord)
	        Enter        ->   CopyMode(MoveToStartOfNextLine)
	        Escape       ->   CopyMode(Close)
	        F            ->   CopyMode(JumpBackward { prev_char: false })
	SHIFT   F            ->   CopyMode(JumpBackward { prev_char: false })
	CTRL    u            ->   CopyMode(ClearPattern)

Key Table: search_mode
----------------------

	        Enter       ->   CopyMode(PriorMatch)
	        Escape      ->   CopyMode(Close)
	CTRL   n           ->   CopyMode(NextMatch)

Mouse
-----

	               Down { streak: 1, button: Left }           ->   SelectTextAtMouseCursor(Cell)
	SHIFT          Down { streak: 1, button: Left }           ->   ExtendSelectionToMouseCursor(Cell)
	SHIFT | ALT    Down { streak: 1, button: Left }           ->   ExtendSelectionToMouseCursor(Block)
	               Drag { streak: 1, button: Left }           ->   ExtendSelectionToMouseCursor(Cell)

Mouse: alt_screen
-----------------

	               Down { streak: 1, button: Left }     ->   SelectTextAtMouseCursor(Cell)
	SHIFT          Down { streak: 1, button: Left }     ->   ExtendSelectionToMouseCursor(Cell)
`

func TestParse(t *testing.T) {
	result := Parse(testInput)

	t.Run("Leader", func(t *testing.T) {
		if result.Leader == nil {
			t.Fatal("expected Leader to be parsed")
		}
		assertEqual(t, "Key", result.Leader.Key, "Char('a')")
		assertEqual(t, "Mods", result.Leader.Mods, "CTRL")
		assertEqual(t, "Timeout", result.Leader.Timeout, "2.001s")
	})

	t.Run("Tables", func(t *testing.T) {
		expected := []string{"Default", "copy_mode", "search_mode", "Mouse", "Mouse: alt_screen"}
		if len(result.Tables) != len(expected) {
			t.Fatalf("expected %d tables, got %d: %v", len(expected), len(result.Tables), result.Tables)
		}
		for i, e := range expected {
			if result.Tables[i] != e {
				t.Errorf("table[%d]: expected %q, got %q", i, e, result.Tables[i])
			}
		}
	})

	t.Run("BindingCount", func(t *testing.T) {
		expected := 23
		if len(result.Bindings) != expected {
			t.Errorf("expected %d bindings, got %d", expected, len(result.Bindings))
		}
	})

	t.Run("DefaultTableBindings", func(t *testing.T) {
		b := result.Bindings[0]
		assertEqual(t, "Table", b.Table, "Default")
		assertEqual(t, "Modifiers", b.Modifiers, "CTRL")
		assertEqual(t, "Key", b.Key, "Tab")
		assertEqual(t, "Action", b.Action, "ActivateTabRelative(1)")
	})

	t.Run("MultipleModifiers", func(t *testing.T) {
		b := result.Bindings[1]
		assertEqual(t, "Modifiers", b.Modifiers, "SHIFT | CTRL")
		assertEqual(t, "Key", b.Key, "Tab")
	})

	t.Run("ThreeModifiers", func(t *testing.T) {
		b := result.Bindings[4]
		assertEqual(t, "Modifiers", b.Modifiers, "SHIFT | ALT | CTRL")
		assertEqual(t, "Key", b.Key, "DownArrow")
		assertEqual(t, "Action", b.Action, "AdjustPaneSize(Down, 1)")
	})

	t.Run("NoModifiers", func(t *testing.T) {
		b := result.Bindings[5]
		assertEqual(t, "Modifiers", b.Modifiers, "")
		assertEqual(t, "Key", b.Key, "Copy")
	})

	t.Run("CopyModeBindings", func(t *testing.T) {
		b := result.Bindings[7]
		assertEqual(t, "Table", b.Table, "copy_mode")
		assertEqual(t, "Modifiers", b.Modifiers, "")
		assertEqual(t, "Key", b.Key, "Tab")
		assertEqual(t, "Action", b.Action, "CopyMode(MoveForwardWord)")
	})

	t.Run("CopyModeSingleCharKey", func(t *testing.T) {
		b := result.Bindings[11]
		assertEqual(t, "Table", b.Table, "copy_mode")
		assertEqual(t, "Modifiers", b.Modifiers, "")
		assertEqual(t, "Key", b.Key, "F")
		assertEqual(t, "Action", b.Action, "CopyMode(JumpBackward { prev_char: false })")
	})

	t.Run("CopyModeWithModifier", func(t *testing.T) {
		b := result.Bindings[12]
		assertEqual(t, "Modifiers", b.Modifiers, "SHIFT")
		assertEqual(t, "Key", b.Key, "F")
	})

	t.Run("SearchModeBindings", func(t *testing.T) {
		b := result.Bindings[14]
		assertEqual(t, "Table", b.Table, "search_mode")
	})

	t.Run("MouseComplexKey", func(t *testing.T) {
		b := result.Bindings[17]
		assertEqual(t, "Table", b.Table, "Mouse")
		assertEqual(t, "Modifiers", b.Modifiers, "")
		assertEqual(t, "Key", b.Key, "Down { streak: 1, button: Left }")
		assertEqual(t, "Action", b.Action, "SelectTextAtMouseCursor(Cell)")
	})

	t.Run("MouseWithModifiers", func(t *testing.T) {
		b := result.Bindings[19]
		assertEqual(t, "Table", b.Table, "Mouse")
		assertEqual(t, "Modifiers", b.Modifiers, "SHIFT | ALT")
		assertEqual(t, "Key", b.Key, "Down { streak: 1, button: Left }")
	})

	t.Run("MouseAltScreenBindings", func(t *testing.T) {
		b := result.Bindings[21]
		assertEqual(t, "Table", b.Table, "Mouse: alt_screen")
	})
}

func TestParseNoLeader(t *testing.T) {
	input := `Default key table
-----------------

	CTRL   Tab   ->   ActivateTabRelative(1)
`
	result := Parse(input)
	if result.Leader != nil {
		t.Error("expected Leader to be nil")
	}
	if len(result.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(result.Bindings))
	}
}

func TestParseLeaderNone(t *testing.T) {
	input := `Leader: Char('a') NONE 1.000s
Default key table
-----------------

	CTRL   Tab   ->   ActivateTabRelative(1)
`
	result := Parse(input)
	if result.Leader == nil {
		t.Fatal("expected Leader to be parsed")
	}
	assertEqual(t, "Mods", result.Leader.Mods, "")
}

func assertEqual(t *testing.T, name, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", name, got, want)
	}
}
