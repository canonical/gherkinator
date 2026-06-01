package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "gherkinator",
	Short: "A testing plan management and generation tool",
}

// setupCommands registers all subcommands on rootCmd.
// It is called by main() and can be called by tests.
func setupCommands() {
	// Reset commands for idempotent setup (useful in tests)
	rootCmd.ResetCommands()

	// 1. INIT COMMAND
	var initCmd = &cobra.Command{
		Use:   "init [directory-name]",
		Short: "Initialize a new test plan directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dirName := args[0]
			if err := os.MkdirAll(dirName, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			emptyPlan := `---
feature: ""
type: ""
status: ""
description: ""
scenarios:
  - ""
examples: []
`
			filePath := filepath.Join(dirName, "test-plan.yaml")
			if err := os.WriteFile(filePath, []byte(emptyPlan), 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			fmt.Printf("Successfully initialized '%s'\n", filePath)
			return nil
		},
	}

	// 2. GENERATE COMMAND
	var outputDir string
	var format string
	var inputFile string
	var generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate gherkin (gh) or markdown (md) files",
		RunE: func(cmd *cobra.Command, args []string) error {
			if format != "gh" && format != "md" {
				return fmt.Errorf("--format must be either 'gh' or 'md'")
			}
			if err := ProcessFile(inputFile, format, outputDir); err != nil {
				return fmt.Errorf("generation failed: %w", err)
			}
			fmt.Println("Generation complete.")
			return nil
		},
	}
	generateCmd.Flags().StringVarP(&outputDir, "output-dir", "o", ".", "Directory to save generated files")
	generateCmd.Flags().StringVar(&format, "format", "gh", "Output format (gh or md)")
	generateCmd.Flags().StringVarP(&inputFile, "input", "i", "test-plan.yaml", "Input YAML file")

	// 3. SERVE COMMAND
	var serveInputFile string
	var serveName string
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Serve test plan docs and watch for changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Derive project name: --name flag, or the directory name of the input file
			projectName := serveName
			if projectName == "" {
				absInput, err := filepath.Abs(serveInputFile)
				if err != nil {
					return fmt.Errorf("failed to resolve input path: %w", err)
				}
				projectName = filepath.Base(filepath.Dir(absInput))
			}

			tmpDir := "./.gherkindocs"
			if err := os.RemoveAll(tmpDir); err != nil {
				return fmt.Errorf("failed to remove existing serve directory: %w", err)
			}

			// 1. Clone slim Sphinx starter pack
			gitCmd := exec.Command(viper.GetString("tools.git"), "clone",
				"https://github.com/canonical/slim-sphinx-docs-starter-pack.git", tmpDir)
			gitCmd.Stdout = os.Stdout
			gitCmd.Stderr = os.Stderr
			if err := gitCmd.Run(); err != nil {
				return fmt.Errorf("failed to clone sphinx starter pack: %w", err)
			}

			// 2-4. Generate markdown, build toctree index, update conf.py
			if err := PrepareSphinxSite(serveInputFile, tmpDir, projectName); err != nil {
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
							if err := PrepareSphinxSite(serveInputFile, tmpDir, projectName); err != nil {
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
			if err := watcher.Add(serveInputFile); err != nil {
				return fmt.Errorf("failed to watch %s: %w", serveInputFile, err)
			}

			// 5. Run make run inside a Bubbletea TUI for clean Ctrl+C handling
			makeBin := viper.GetString("tools.make")
			env := os.Environ()
			env = append(env, fmt.Sprintf("PYTHON_BIN=%s", viper.GetString("tools.python3")))
			env = append(env, fmt.Sprintf("PIP_BIN=%s", viper.GetString("tools.pip")))

			p := tea.NewProgram(initialServeModel(makeBin, docsDir, env))
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("serve TUI error: %w", err)
			}
			return nil
		},
	}
	serveCmd.Flags().StringVarP(&serveInputFile, "input", "i", "test-plan.yaml", "Input YAML file")
	serveCmd.Flags().StringVarP(&serveName, "name", "n", "", "Project name for the documentation (defaults to input directory name)")

	// 5. DELETE COMMAND
	var skipConfirm bool
	var deleteInputFile string
	var deleteCmd = &cobra.Command{
		Use:   "delete [feature-names...]",
		Short: "Delete test plans by feature name from test-plan.yaml",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			plans, err := LoadTestPlans(deleteInputFile)
			if err != nil {
				return fmt.Errorf("failed to load test plans: %w", err)
			}

			remaining, deleted := DeleteTestPlans(plans, args)
			if len(deleted) == 0 {
				fmt.Println("No matching test plans found.")
				return nil
			}

			if !skipConfirm {
				if !ConfirmDeletion(deleted, os.Stdin) {
					fmt.Println("Delete aborted.")
					return nil
				}
			}

			if err := WriteTestPlans(deleteInputFile, remaining); err != nil {
				return fmt.Errorf("failed to write updated test plans: %w", err)
			}

			fmt.Printf("Successfully deleted %d test plan(s).\n", len(deleted))
			return nil
		},
	}
	deleteCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")
	deleteCmd.Flags().StringVarP(&deleteInputFile, "input", "i", "test-plan.yaml", "Input YAML file")

	// 6. CLEAN COMMAND
	var cleanDir string
	var cleanCmd = &cobra.Command{
		Use:   "clean",
		Short: "Clean generated files and temporary directories from the test plan directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := CleanDirectory(cleanDir); err != nil {
				return fmt.Errorf("clean failed: %w", err)
			}
			fmt.Println("Clean complete.")
			return nil
		},
	}
	cleanCmd.Flags().StringVarP(&cleanDir, "dir", "d", ".", "Directory to clean")

	rootCmd.AddCommand(initCmd, generateCmd, serveCmd, deleteCmd, cleanCmd)
}

func main() {
	cobra.OnInitialize(initConfig)
	setupCommands()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
