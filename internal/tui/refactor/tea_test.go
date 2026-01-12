package refactor

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/aymanbagabas/go-udiff"
	tea "github.com/charmbracelet/bubbletea"
)

var update = flag.Bool("update", false, "update golden files")

type component interface {
	Update(msg tea.Msg) tea.Cmd
}

func sendAndWait(c component, msg tea.Msg) {
	msgs := []tea.Msg{msg}
	for len(msgs) > 0 {
		msg, msgs = msgs[0], msgs[1:]
		if cmd := c.Update(msg); cmd != nil {
			msg := cmd()
			switch msg.(type) {
			case tea.BatchMsg:
				for _, c := range msg.(tea.BatchMsg) {
					if m := c(); m != nil {
						msgs = append(msgs, m)
					}
				}
			default:
				msgs = append(msgs, msg)
			}
		}
	}
}

// requireEqual is borrowed from teatest's golden package
func requireEqual[T []byte | string](tb testing.TB, out T) {
	tb.Helper()

	golden := filepath.Join("testdata", tb.Name()+".golden")
	if *update {
		if err := os.MkdirAll(filepath.Dir(golden), 0o750); err != nil { //nolint: mnd
			tb.Fatal(err)
		}
		if err := os.WriteFile(golden, []byte(out), 0o600); err != nil { //nolint: mnd
			tb.Fatal(err)
		}
	}

	goldenBts, err := os.ReadFile(golden)
	if err != nil {
		tb.Fatal(err)
	}

	goldenStr := string(goldenBts)
	outStr := string(out)

	diff := udiff.Unified("golden", "run", goldenStr, outStr)
	if diff != "" {
		tb.Fatalf("output does not match, expected:\n\n%s\n\ngot:\n\n%s\n\ndiff:\n\n%s", goldenStr, outStr, diff)
	}
}
