package serve

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

const maxServeLogLines = 100

// serveModel is the Bubbletea model for the gherkinator serve command.
// It starts make run as a child process, streams its combined output
// into a log view, and shuts down cleanly on Ctrl+C.
type serveModel struct {
	makeBin  string
	docsDir  string
	env      []string
	lines    []string
	cmd      *exec.Cmd
	outputCh <-chan string
	doneCh   <-chan error
	quitting bool
	err      error
}

// serveLine is a Bubbletea message carrying a single line of output.
type serveLine string

// serveExited is a Bubbletea message sent when the child process exits.
type serveExited struct{ err error }

// InitialServeModel returns a Bubbletea model ready to be passed to
// tea.NewProgram. The model is inert until Init() is called by the
// Bubbletea runtime.
func InitialServeModel(makeBin string, docsDir string, env []string) serveModel {
	return serveModel{
		makeBin: makeBin,
		docsDir: docsDir,
		env:     env,
	}
}

func (m serveModel) Init() tea.Cmd {
	return func() tea.Msg {
		return serveStartMsg{}
	}
}

// serveStartMsg triggers the process launch from inside the Update loop
// so we can store the cmd and channels on the model.
type serveStartMsg struct{}

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
	case serveStartMsg:
		cmd := exec.Command(m.makeBin, "run")
		cmd.Dir = m.docsDir
		cmd.Env = m.env

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			m.quitting = true
			m.err = err
			return m, tea.Quit
		}
		cmd.Stderr = cmd.Stdout // merge stderr into stdout

		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		if err := cmd.Start(); err != nil {
			m.quitting = true
			m.err = err
			return m, tea.Quit
		}

		outputCh := make(chan string, 64)
		doneCh := make(chan error, 1)
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				outputCh <- scanner.Text()
			}
			close(outputCh)
			doneCh <- cmd.Wait()
		}()

		m.cmd = cmd
		m.outputCh = outputCh
		m.doneCh = doneCh
		return m, waitForOutput(outputCh, doneCh)

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
