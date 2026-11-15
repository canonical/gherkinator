package serve

import (
	"bufio"
	"cmp"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/canonical/gherkinator/internal/common"
)

// ensureWritableDir verifies that dir exists and accepts writes by creating
// and removing a probe file inside it. The probe avoids relying on the file
// mode bits alone, which do not reflect mounts or confinement rules.
func ensureWritableDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("serve directory %s cannot be created: %w", dir, err)
	}

	probe, err := os.CreateTemp(dir, ".gherkinator-probe-*")
	if err != nil {
		return fmt.Errorf("serve directory %s is not writable: %w", dir, err)
	}
	probeName := probe.Name()
	if err := probe.Close(); err != nil {
		_ = os.Remove(probeName)
		return fmt.Errorf("failed to close probe file in serve directory %s: %w", dir, err)
	}
	if err := os.Remove(probeName); err != nil {
		return fmt.Errorf("failed to remove probe file in serve directory %s: %w", dir, err)
	}
	return nil
}

// prepareServeSite resolves the staging directory, verifies it is usable,
// clones the sphinx-stack template into it, and generates the initial
// documentation set. It returns the docs directory to serve and the
// regeneration closure used for live reloads.
func prepareServeSite(cwd string, inputFiles []string, projectName string, opts common.FilterOptions) (string, func() error, error) {
	if projectName == "" {
		projectName = filepath.Base(cwd)
	}

	tmpDir, err := common.StagingDirFor(cwd)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve serve directory: %w", err)
	}

	// Verify the serve directory is usable before doing any work so that
	// misconfiguration (e.g. a read-only path inside the snap) produces an
	// actionable error instead of a cryptic git failure.
	if err := ensureWritableDir(tmpDir); err != nil {
		return "", nil, fmt.Errorf("%w (configure an alternative with the GHERKINATOR_SERVE_DIR environment variable or the serve.dir config key)", err)
	}

	// Wipe any previous staging content. Failures are non-fatal because the
	// clone below reports an actionable error when leftovers remain.
	if err := os.RemoveAll(tmpDir); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not fully clear serve directory %s: %v\n", tmpDir, err) //nolint:errcheck // Writing to stderr; error is not actionable
	}

	// Clone Canonical Sphinx stack
	gitCmd := exec.Command(viper.GetString("tools.git"), "clone",
		"https://github.com/canonical/sphinx-stack.git", tmpDir)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return "", nil, fmt.Errorf("failed to clone sphinx-stack into %s (it may still contain leftovers; remove them manually or point GHERKINATOR_SERVE_DIR elsewhere): %w", tmpDir, err)
	}

	// Prune sphinx-stack template content and default index.rst
	docsDir := filepath.Join(tmpDir, "docs")
	if err := PruneSphinxStackDefaults(docsDir); err != nil {
		return "", nil, fmt.Errorf("failed to prune sphinx-stack defaults: %w", err)
	}

	// regenerateDocs clears generated type subdirs, regenerates docs from
	// every discovered input file, and rebuilds the Sphinx index.
	regenerateDocs := func() error {
		if err := CleanGeneratedDocs(docsDir); err != nil {
			return fmt.Errorf("failed to clean generated docs: %w", err)
		}
		var merged []common.TestPlan
		for _, file := range inputFiles {
			plans, err := GenerateSphinxDocs(file, docsDir, opts)
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
		return "", nil, fmt.Errorf("failed to prepare sphinx site: %w", err)
	}
	return docsDir, regenerateDocs, nil
}

// servingURL returns the web address sphinx-autobuild serves the docs on.
// The sphinx-stack Makefile pins the host and port through SPHINX_HOST and
// SPHINX_PORT (defaulting to 127.0.0.1:8000), and the serve command passes
// the environment through to make, so the same variables are honored here.
func servingURL() string {
	host := cmp.Or(os.Getenv("SPHINX_HOST"), "127.0.0.1")
	port := cmp.Or(os.Getenv("SPHINX_PORT"), "8000")
	return fmt.Sprintf("http://%s:%s", host, port)
}

// checkServingPortFree verifies that nothing is listening on the address
// the sphinx server will bind, so a conflict surfaces before the
// multi-second staging build instead of after it. A successful connection
// means the port is taken; any other dial error means the port is free.
func checkServingPortFree(url string) error {
	conn, err := net.DialTimeout("tcp", strings.TrimPrefix(url, "http://"), time.Second)
	if err != nil {
		return nil
	}
	_ = conn.Close()
	return fmt.Errorf("port is already in use: %s (stop the process listening there or pick another port via SPHINX_PORT)", url)
}

// Run is the high-level orchestrator for the serve command. It clones the
// Canonical Sphinx stack, generates initial documentation from the
// provided input files, sets up an fsnotify watcher for live reload, and
// launches `make run` inside a Bubbletea TUI.
//
// projectName controls the title of the rendered Sphinx site; pass "" to
// use the basename of the current working directory.
//
// The filters in opts are intersected (see GenerateSphinxDocs); pass the
// zero value to render every plan.
func Run(inputFiles []string, projectName string, opts common.FilterOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to resolve working directory: %w", err)
	}

	// The kernel-managed lock serializes serve invocations and outlives
	// nothing: it is released automatically on any exit, including SIGKILL.
	releaseLock, err := lockServeInstance()
	if err != nil {
		return err
	}
	defer releaseLock()

	url := servingURL()
	if err := checkServingPortFree(url); err != nil {
		return err
	}

	docsDir, regenerateDocs, err := prepareServeSite(cwd, inputFiles, projectName, opts)
	if err != nil {
		return err
	}

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

	// Start make run as a process-group leader so the whole sphinx tree can
	// be terminated as a unit. The TUI only renders its output; the serve
	// command owns the child and performs cleanup on every exit path.
	cmd := exec.Command(viper.GetString("tools.make"), "run")
	cmd.Dir = docsDir
	cmd.Env = os.Environ()
	// PYTHONUNBUFFERED keeps sphinx-autobuild's own status lines ([sphinx-autobuild]
	// banners) flowing through the output pipe: with piped stdout, Python would
	// otherwise block-buffer them and they would never reach the log view.
	cmd.Env = append(cmd.Env,
		"PYTHONUNBUFFERED=1",
		fmt.Sprintf("PYTHON_BIN=%s", viper.GetString("tools.python3")),
		fmt.Sprintf("PIP_BIN=%s", viper.GetString("tools.pip")),
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create docs output pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s run in %s: %w", cmd.Path, docsDir, err)
	}
	pgid := -cmd.Process.Pid

	// sphinx-autobuild suppresses the uvicorn banner that announces the
	// web address, so surface it before and during the TUI.
	fmt.Printf("Serving documentation at %s\n", url) //nolint:errcheck // Writing to stdout; error is not actionable

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

	// Terminate the child tree on signals the TUI does not handle (e.g.
	// SIGTERM from `timeout` or `kill`), so no orphaned sphinx server is
	// left serving a doomed staging directory.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(sigCh)
	go func() {
		sig := <-sigCh
		_ = syscall.Kill(pgid, syscall.SIGKILL)
		_ = sig
	}()

	defer func() {
		// Idempotent: a no-op when the TUI or the signal handler already
		// terminated the process group.
		_ = syscall.Kill(pgid, syscall.SIGKILL)
		select {
		case err := <-doneCh:
			_ = err
		case <-time.After(5 * time.Second):
		}
	}()

	p := tea.NewProgram(InitialServeModel(cmd, outputCh, doneCh, url))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("serve TUI error: %w", err)
	}
	return nil
}
