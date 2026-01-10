package pane

import (
	"maps"
	"reflect"
	"slices"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var testPanes = map[Name]tea.Model{
	"A": testPane{"A", nil, 0},
	"B": testPane{"B", func() tea.Msg {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
	}, 0},
	"C": testPane{"C", func() tea.Msg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}} }, 0},
}

func TestManager_Init(t *testing.T) {
	m := makeManager(testPanes)
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected init command, got nil")
	}
	msgs := batchCmdToMsgs(t, cmd)
	slices.SortFunc(msgs, func(a, b tea.Msg) int {
		return slices.Compare(a.(tea.KeyMsg).Runes, b.(tea.KeyMsg).Runes)
	})
	want := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}},
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}},
	}
	if !reflect.DeepEqual(msgs, want) {
		t.Errorf("expected key runes to be sorted, got %v", msgs)
	}
}

func TestManager_Update(t *testing.T) {
	t.Run("key input is sent to active pane", func(t *testing.T) {
		m := makeManager(testPanes)
		m, _ = m.Update(ActivateMsg{Pane: "B"})
		m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		if cmd != nil {
			t.Fatalf("expected no update command, got %v", cmd())
		}
		if m.(Manager).panes["A"].(testPane).calls != 0 {
			t.Errorf("expected pane A to have not been updated, got %d", m.(Manager).panes["A"].(testPane).calls)
		}
		if m.(Manager).panes["B"].(testPane).calls != 1 {
			t.Errorf("expected pane B to have been updated, got %d", m.(Manager).panes["B"].(testPane).calls)
		}
	})
	t.Run("non-key input is broadcast to all panes", func(t *testing.T) {
		m := makeManager(testPanes)
		m, _ = m.Update(tea.WindowSizeMsg{})
		for paneName := range testPanes {
			if m.(Manager).panes[paneName].(testPane).calls == 0 {
				t.Errorf("expected pane %s to have been updated, got %d", paneName, m.(Manager).panes[paneName].(testPane).calls)
			}
		}
	})
}

func TestManager_View(t *testing.T) {
	var m tea.Model = New(testPanes)
	// not initialized: no output
	if got := m.View(); got != "" {
		t.Fatalf("expected no view, got %s", got)
	}
	// active pane is output
	for paneName := range testPanes {
		m, _ = m.Update(ActivateMsg{Pane: paneName})
		if got := m.View(); got != string(paneName) {
			t.Fatalf("expected view from %s', got %s", paneName, got)
		}
	}
}

func makeManager(panes map[Name]tea.Model) tea.Model {
	tpc := make(map[Name]tea.Model)
	maps.Copy(tpc, panes)
	return New(tpc)
}

func batchCmdToMsgs(t testing.TB, cmd tea.Cmd) []tea.Msg {
	t.Helper()
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected batch message, got %T", cmd())
	}
	msgs := make([]tea.Msg, len(batch))
	for i, cmd := range batch {
		msgs[i] = cmd()
	}
	return msgs
}

var _ tea.Model = &testPane{}

type testPane struct {
	name  Name
	init  tea.Cmd
	calls int
}

func (p testPane) Init() tea.Cmd {
	return p.init
}

func (p testPane) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	p.calls++
	return p, nil
}

func (p testPane) View() string {
	return string(p.name)
}
