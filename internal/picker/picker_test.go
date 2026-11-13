package picker

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func sampleOptions() []Option {
	return []Option{
		{Key: "raycast", Label: "Raycast"},
		{Key: "figma", Label: "Figma"},
		{Key: "zed", Label: "Zed"},
	}
}

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func send(m Model, msgs ...tea.Msg) Model {
	for _, msg := range msgs {
		upd, _ := m.Update(msg)
		m = upd.(Model)
	}
	return m
}

func TestPicker_InitialSelection(t *testing.T) {
	m := New("title", sampleOptions(), []string{"figma"})
	if got := m.Selected(); !reflect.DeepEqual(got, []string{"figma"}) {
		t.Fatalf("initial Selected() = %v, want [figma]", got)
	}
}

func TestPicker_ToggleAndConfirm(t *testing.T) {
	m := New("title", sampleOptions(), nil)
	// cursor at 0 (raycast), toggle → select raycast
	m = send(m, keyMsg(" "))
	// down to figma, toggle → select figma
	m = send(m, keyMsg("down"), keyMsg(" "))
	// down to zed, leave unselected
	m = send(m, keyMsg("down"))
	// enter → confirm
	m = send(m, keyMsg("enter"))
	if !m.Confirmed() {
		t.Fatal("Confirmed() false after enter")
	}
	if m.Cancelled() {
		t.Fatal("Cancelled() true after enter")
	}
	want := []string{"raycast", "figma"}
	if got := m.Selected(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Selected() = %v, want %v", got, want)
	}
}

func TestPicker_EscCancels(t *testing.T) {
	m := New("title", sampleOptions(), []string{"raycast"})
	m = send(m, keyMsg(" ")) // toggle off
	m = send(m, keyMsg("esc"))
	if !m.Cancelled() {
		t.Fatal("Cancelled() false after esc")
	}
	if m.Confirmed() {
		t.Fatal("Confirmed() true after esc")
	}
}

func TestPicker_SelectAllNone(t *testing.T) {
	m := New("title", sampleOptions(), nil)
	m = send(m, keyMsg("a"))
	if got := len(m.Selected()); got != 3 {
		t.Fatalf("after 'a' got %d selected, want 3", got)
	}
	m = send(m, keyMsg("n"))
	if got := len(m.Selected()); got != 0 {
		t.Fatalf("after 'n' got %d selected, want 0", got)
	}
}

func TestPicker_CursorBounds(t *testing.T) {
	m := New("title", sampleOptions(), nil)
	// up at top should stay at 0
	m = send(m, keyMsg("up"))
	// many downs past the bottom should clamp at last index
	for i := 0; i < 10; i++ {
		m = send(m, keyMsg("down"))
	}
	// toggle the last option
	m = send(m, keyMsg(" "), keyMsg("enter"))
	want := []string{"zed"}
	if got := m.Selected(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Selected() = %v, want %v", got, want)
	}
}
