package serve

import (
	"os/exec"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestInitialServeModel_ViewShowsServingURL(t *testing.T) {
	// sphinx-autobuild suppresses uvicorn's INFO banner announcing the web
	// address, so the TUI header must carry the URL in every frame.
	cmd := exec.Command("true")
	model := InitialServeModel(cmd, nil, nil, "http://127.0.0.1:8000")

	view := model.View()
	assert.Contains(t, view, "Documentation: http://127.0.0.1:8000")
}

func TestInitialServeModel_ViewWithoutURL(t *testing.T) {
	// An empty URL must not render a dangling header line.
	cmd := exec.Command("true")
	model := InitialServeModel(cmd, nil, nil, "")

	view := model.View()
	assert.NotContains(t, view, "Documentation:")
	assert.Contains(t, view, "press Ctrl+C to stop")
}

func TestServeModel_ViewKeepsURLAfterLogLines(t *testing.T) {
	// The URL must persist in the header even when the log view fills with
	// build output, because that output is what pushes the URL out of view.
	outputCh := make(chan string, 8)
	doneCh := make(chan error, 1)
	var current tea.Model = InitialServeModel(exec.Command("true"), outputCh, doneCh, "http://127.0.0.1:8000")
	for range 150 {
		current, _ = current.Update(serveLine(strings.Repeat("build line ", 10)))
	}

	view := current.View()
	assert.Contains(t, view, "Documentation: http://127.0.0.1:8000")
	assert.Contains(t, view, "build line")
}

func TestServingURL_Defaults(t *testing.T) {
	t.Setenv("SPHINX_HOST", "")
	t.Setenv("SPHINX_PORT", "")

	assert.Equal(t, "http://127.0.0.1:8000", servingURL())
}

func TestServingURL_EnvOverride(t *testing.T) {
	// The sphinx-stack Makefile honors SPHINX_HOST/SPHINX_PORT from the
	// environment (its ?= assignments), and serve passes the environment
	// through to make, so the displayed URL must track them.
	t.Setenv("SPHINX_HOST", "0.0.0.0")
	t.Setenv("SPHINX_PORT", "9001")

	assert.Equal(t, "http://0.0.0.0:9001", servingURL())
}
