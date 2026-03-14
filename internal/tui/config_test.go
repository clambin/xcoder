package tui

import (
	"testing"

	"github.com/charmbracelet/x/exp/golden"
	"github.com/clambin/xcoder/internal/pipeline"
)

func TestConfigView_View(t *testing.T) {
	cfg := pipeline.Configuration{
		Input:       "/a/very/long/path/that/will/be/truncated",
		ProfileName: "foo",
		Active:      false,
		Remove:      true,
		Overwrite:   true,
	}
	golden.RequireEqual(t, newConfigView(cfg, ConfigStyles{}))
}
