package serve

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"gherkinator/internal/common"
)

// Run is the high-level orchestrator for the serve command. It clones the
// slim Sphinx starter pack, generates initial documentation from the
// provided input files, sets up an fsnotify watcher for live reload, and
// launches `make run` inside a Bubbletea TUI.
//
// projectName controls the title of the rendered Sphinx site; pass "" to
// use the basename of the current working directory.
//
// riskFilter and statusFilter are intersected (see GenerateSphinxDocs).
func Run(inputFiles []string, projectName string, riskFilter string, statusFilter string) error {
	if projectName == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to resolve working directory: %w", err)
		}
		projectName = filepath.Base(cwd)
	}

	tmpDir := "./.gherkindocs"
	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("failed to remove existing serve directory: %w", err)
	}

	// Clone slim Sphinx starter pack
	gitCmd := exec.Command(viper.GetString("tools.git"), "clone",
		"https://github.com/canonical/slim-sphinx-docs-starter-pack.git", tmpDir)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone sphinx starter pack: %w", err)
	}

	// regenerateDocs clears generated type subdirs, regenerates docs from
	// every discovered input file, and rebuilds the Sphinx index.
	regenerateDocs := func() error {
		docsDir := filepath.Join(tmpDir, "docs")
		if err := CleanGeneratedDocs(docsDir); err != nil {
			return fmt.Errorf("failed to clean generated docs: %w", err)
		}
		var merged []common.TestPlan
		for _, file := range inputFiles {
			plans, err := GenerateSphinxDocs(file, docsDir, riskFilter, statusFilter)
			if err != nil {
				return fmt.Errorf("failed to generate sphinx docs for %s: %w", file, err)
			}
			merged = append(merged, plans...)
		}
		if err := BuildSphinxIndex(docsDir, merged); err != nil {
			return fmt.Errorf("failed to build sphinx index: %w", err)
		}
		confPath := filepath.Join(docsDir, "conf.py")
		if err := UpdateConfPy(confPath, projectName); err != nil {
			return fmt.Errorf("failed to update conf.py: %w", err)
		}
		return nil
	}

	if err := regenerateDocs(); err != nil {
		return fmt.Errorf("failed to prepare sphinx site: %w", err)
	}

	docsDir := filepath.Join(tmpDir, "docs")

	// fsnotify Watcher for Live Reloading
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer func() {
		_ = watcher.Close()
	}()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					fmt.Println("Change detected. Rebuilding docs...")
					if err := regenerateDocs(); err != nil {
						//nolint:errcheck // Writing to stderr; error is not actionable
						fmt.Fprintf(os.Stderr, "Rebuild error: %v\n", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				//nolint:errcheck // Writing to stderr; error is not actionable
				fmt.Fprintf(os.Stderr, "Watcher error: %v\n", err)
			}
		}
	}()
	for _, file := range inputFiles {
		if err := watcher.Add(file); err != nil {
			return fmt.Errorf("failed to watch %s: %w", file, err)
		}
	}

	// Run make run inside a Bubbletea TUI for clean Ctrl+C handling
	makeBin := viper.GetString("tools.make")
	env := os.Environ()
	env = append(env, fmt.Sprintf("PYTHON_BIN=%s", viper.GetString("tools.python3")))
	env = append(env, fmt.Sprintf("PIP_BIN=%s", viper.GetString("tools.pip")))

	p := tea.NewProgram(InitialServeModel(makeBin, docsDir, env))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("serve TUI error: %w", err)
	}
	return nil
}
