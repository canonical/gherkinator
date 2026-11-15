package serve

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

const maxServeLogLines = 100

// serveModel is the Bubbletea model for the gherkinator serve command.
// It streams the output of the already-running `make run` child process
// into a log view and shuts down cleanly on Ctrl+C. The child process is
// owned by the serve command (see Run), which handles signal-driven
// cleanup; the model only mirrors its state.
type serveModel struct {
	cmd        *exec.Cmd
	servingURL string
	lines      []string
	outputCh   <-chan string
	doneCh     <-chan error
	quitting   bool
	err        error
}

// serveLine is a Bubbletea message carrying a single line of output.
type serveLine string

// serveExited is a Bubbletea message sent when the child process exits.
type serveExited struct{ err error }

// InitialServeModel returns a Bubbletea model ready to be passed to
// tea.NewProgram. cmd must already be started, with outputCh and doneCh
// wired to its streams. servingURL is displayed in the header because
// sphinx-autobuild suppresses the uvicorn banner that normally announces
// the web address.
func InitialServeModel(cmd *exec.Cmd, outputCh <-chan string, doneCh <-chan error, servingURL string) serveModel {
	return serveModel{
		cmd:        cmd,
		outputCh:   outputCh,
		doneCh:     doneCh,
		servingURL: servingURL,
	}
}

func (m serveModel) Init() tea.Cmd {
	return waitForOutput(m.outputCh, m.doneCh)
}

// waitForOutput returns a Cmd that reads one item from either the output
// channel or the done channel and returns the appropriate message.
func waitForOutput(outputCh <-chan string, doneCh <-chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case line, ok := <-outputCh:
			if !ok {
				// Channel closed, wait for process exit.
				err := <-doneCh
				return serveExited{err: err}
			}
			return serveLine(line)
		case err := <-doneCh:
			return serveExited{err: err}
		}
	}
}

func (m serveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			if m.cmd != nil && m.cmd.Process != nil {
				_ = syscall.Kill(-m.cmd.Process.Pid, syscall.SIGKILL)
			}
			return m, tea.Quit
		}

	case serveLine:
		m.lines = append(m.lines, string(msg))
		if len(m.lines) > maxServeLogLines {
			m.lines = m.lines[len(m.lines)-maxServeLogLines:]
		}
		return m, waitForOutput(m.outputCh, m.doneCh)

	case serveExited:
		m.quitting = true
		m.err = msg.err
		return m, tea.Quit
	}
	return m, nil
}

func (m serveModel) View() string {
	var sb strings.Builder
	sb.WriteString("gherkinator serve \u2014 press Ctrl+C to stop\n")
	if m.servingURL != "" {
		fmt.Fprintf(&sb, "Documentation: %s\n", m.servingURL)
	}
	sb.WriteString(strings.Repeat("\u2500", 50) + "\n")

	for _, line := range m.lines {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	if m.quitting {
		if m.err != nil {
			fmt.Fprintf(&sb, "\nProcess exited with error: %v\n", m.err)
		} else {
			sb.WriteString("\nStopped.\n")
		}
	}
	return sb.String()
}
