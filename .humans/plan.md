# `gherkinator` implementation plan

This document outlines the complete plan to build `gherkinator`, a Golang-based CLI tool that manages centralized YAML test plans. It parses, validates, and transpiles these plans into Gherkin feature files or Markdown documentation, provides an interactive local documentation server using the Canonical Sphinx starter pack, and is distributed as a Snap package.

### Phase 1: Dependencies & Setup
Initialize the Go module and install all the required dependencies.

```bash
go mod init gherkinator

# YAML and Gherkin parsing
go get go.yaml.in/yaml/v3
go get [github.com/cucumber/gherkin/go/v26](https://github.com/cucumber/gherkin/go/v26)

# CLI Framework and Configuration Management
go get [github.com/spf13/cobra@latest](https://github.com/spf13/cobra@latest)
go get [github.com/spf13/viper@latest](https://github.com/spf13/viper@latest)

# Interactive Prompts and File Watching
go get [github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
go get [github.com/charmbracelet/bubbles](https://github.com/charmbracelet/bubbles)
go get [github.com/fsnotify/fsnotify](https://github.com/fsnotify/fsnotify)

# Unit Testing
go get [github.com/stretchr/testify](https://github.com/stretchr/testify)
```

### Phase 2: YAML Schema Definition & Validation

Define the data structure mapped to the specification's YAML format and enforce strict validation for allowed test types and implementation statuses.

```go
package main

import "fmt"

// TestPlan represents the YAML schema for a single test plan document.
type TestPlan struct {
	Feature     string      `yaml:"feature"`
	Type        string      `yaml:"type"`
	Status      string      `yaml:"status"`
	Issues      *string     `yaml:"issues,omitempty"`
	Docs        *string     `yaml:"docs,omitempty"`
	Description *string     `yaml:"description,omitempty"`
	Background  *string     `yaml:"background,omitempty"`
	Scenarios   []string    `yaml:"scenarios"`
	Examples    [][]string  `yaml:"examples,omitempty"`
}

// ValidateSchema ensures the test plan uses accepted types and statuses.
func ValidateSchema(plan TestPlan) error {
	validTypes := map[string]bool{
		"functional": true, "solution": true, "performance": true, 
		"reliability": true, "security": true,
	}
	if !validTypes[plan.Type] {
		return fmt.Errorf("invalid type '%s': must be one of 'functional', 'solution', 'performance', 'reliability', or 'security'", plan.Type)
	}

	validStatuses := map[string]bool{
		"planned": true, "implemented": true, "deprecated": true,
	}
	if !validStatuses[plan.Status] {
		return fmt.Errorf("invalid status '%s': must be one of 'planned', 'implemented', or 'deprecated'", plan.Status)
	}
	return nil
}
```

---

### Phase 3: Transpilation Engine (Gherkin & Markdown)
Implement the core functions to convert the validated `TestPlan` struct into standard Gherkin syntax and readable Markdown documentation.

```go
package main

import (
	"fmt"
	"regexp"
	"strings"
)

// GenerateGherkin transpiles a TestPlan struct into a valid Gherkin string.
func GenerateGherkin(plan TestPlan) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Feature: %s\n", plan.Feature))
	if plan.Description != nil { builder.WriteString(fmt.Sprintf("%s\n", *plan.Description)) }
	if plan.Background != nil {
		builder.WriteString("Background:\n")
		builder.WriteString(fmt.Sprintf("%s\n", *plan.Background))
	}
	builder.WriteString(fmt.Sprintf("@%s\n", plan.Type))

	hasExamples := len(plan.Examples) > 0
	for _, scenario := range plan.Scenarios {
		parts := strings.SplitN(scenario, "\n", 2)
		title, steps := strings.TrimSpace(parts[0]), ""
		if len(parts) > 1 { steps = parts[1] }

		if hasExamples {
			builder.WriteString(fmt.Sprintf("Scenario Outline: %s\n", title))
		} else {
			builder.WriteString(fmt.Sprintf("Scenario: %s\n", title))
		}
		builder.WriteString(fmt.Sprintf("%s\n", steps))
	}

	if hasExamples {
		builder.WriteString("Examples:\n")
		paramRegex := regexp.MustCompile(`<([^>]+)>`)
		var params []string
		seen := make(map[string]bool)
		for _, scenario := range plan.Scenarios {
			for _, match := range paramRegex.FindAllStringSubmatch(scenario, -1) {
				if !seen[match[1]] {
					seen[match[1]] = true
					params = append(params, match[1])
				}
			}
		}
		builder.WriteString("| " + strings.Join(params, " | ") + " |\n")
		for _, row := range plan.Examples {
			builder.WriteString("| " + strings.Join(row, " | ") + " |\n")
		}
	}
	return builder.String()
}

// GenerateMarkdown transpiles a TestPlan struct into a Markdown string.
func GenerateMarkdown(plan TestPlan) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# %s\n\n", plan.Feature))
	builder.WriteString(fmt.Sprintf("- **Type:** %s\n", plan.Type))
	builder.WriteString(fmt.Sprintf("- **Status:** %s\n", plan.Status))
	if plan.Issues != nil { builder.WriteString(fmt.Sprintf("- **Issues:** %s\n", *plan.Issues)) }
	if plan.Docs != nil { builder.WriteString(fmt.Sprintf("- **Docs:** %s\n", *plan.Docs)) }
	builder.WriteString("\n")

	if plan.Description != nil {
		builder.WriteString("## Description\n\n")
		builder.WriteString(fmt.Sprintf("%s\n\n", *plan.Description))
	}

	if plan.Background != nil {
		builder.WriteString("## Background\n\n")
		for _, line := range strings.Split(strings.TrimSpace(*plan.Background), "\n") {
			builder.WriteString(fmt.Sprintf("- %s\n", line))
		}
		builder.WriteString("\n")
	}

	hasExamples := len(plan.Examples) > 0
	if hasExamples {
		builder.WriteString("## Scenario Outlines\n\n")
	} else {
		builder.WriteString("## Scenarios\n\n")
	}
	for _, scenario := range plan.Scenarios {
		parts := strings.SplitN(scenario, "\n", 2)
		builder.WriteString(fmt.Sprintf("### %s\n\n", strings.TrimSpace(parts[0])))
		if len(parts) > 1 {
			for _, step := range strings.Split(strings.TrimSpace(parts[1]), "\n") {
				step = strings.Replace(step, "Given ", "**Given** ", 1)
				step = strings.Replace(step, "When ", "**When** ", 1)
				step = strings.Replace(step, "Then ", "**Then** ", 1)
				step = strings.Replace(step, "And ", "**And** ", 1)
				builder.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(step)))
			}
		}
		builder.WriteString("\n")
	}

	if len(plan.Examples) > 0 {
		builder.WriteString("## Examples\n\n")
		paramRegex := regexp.MustCompile(`<([^>]+)>`)
		var params []string
		seen := make(map[string]bool)
		for _, scenario := range plan.Scenarios {
			for _, match := range paramRegex.FindAllStringSubmatch(scenario, -1) {
				if !seen[match[1]] {
					seen[match[1]] = true
					params = append(params, match[1])
				}
			}
		}
		if len(params) > 0 {
			builder.WriteString("| " + strings.Join(params, " | ") + " |\n")
			builder.WriteString("| " + strings.Repeat("--- | ", len(params)) + "\n")
		}
		for _, row := range plan.Examples {
			builder.WriteString("| " + strings.Join(row, " | ") + " |\n")
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
```

---

### Phase 4: Core Processing & Syntax Validation
Read the YAML file (handling multi-document streams separated by `---`), validate the schema, transpile, and guarantee Gherkin validity using the official Cucumber parser.

```go
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	gherkin "[github.com/cucumber/gherkin/go/v26](https://github.com/cucumber/gherkin/go/v26)"
	"go.yaml.in/yaml/v3"
)

func ValidateGherkin(gherkinText string) error {
	reader := strings.NewReader(gherkinText)
	_, err := gherkin.ParseGherkinDocument(reader, (&gherkin.Incrementing{}).NewId)
	return err
}

func ProcessFile(filename string, format string, outputDir string) error {
	file, err := os.Open(filename)
	if err != nil { return err }
	defer file.Close()

	os.MkdirAll(outputDir, 0755)

	decoder := yaml.NewDecoder(file)
	for i := 1; ; i++ {
		var plan TestPlan
		if err := decoder.Decode(&plan); err == io.EOF { break } else if err != nil { return err }

		if err := ValidateSchema(plan); err != nil { return err }

		var output string
		var ext string
		if format == "gh" {
			output = GenerateGherkin(plan)
			if err := ValidateGherkin(output); err != nil { return err }
			ext = ".feature"
		} else if format == "md" {
			output = GenerateMarkdown(plan)
			ext = ".md"
		}

		safeFilename := strings.ReplaceAll(strings.ToLower(plan.Feature), " ", "_")
		if safeFilename == "" { safeFilename = fmt.Sprintf("plan_%d", i) }
		
		outPath := filepath.Join(outputDir, safeFilename+ext)
		os.WriteFile(outPath, []byte(output), 0644)
	}
	return nil
}
```

---

### Phase 5: Viper Configuration Management
Configure Viper to manage executable paths (`git`, `python3`, `pip`, `make`) for the Sphinx environment. Users can override these via `.gherkinator.yaml` or environment variables.

```go
package main

import (
	"[github.com/spf13/viper](https://github.com/spf13/viper)"
)

func initConfig() {
	viper.SetDefault("tools.git", "git")
	viper.SetDefault("tools.python3", "python3")
	viper.SetDefault("tools.pip", "pip")
	viper.SetDefault("tools.make", "make")

	viper.SetConfigName(".gherkinator")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.SetEnvPrefix("gherkinator")
	viper.AutomaticEnv()
	viper.ReadInConfig()
}
```

---

### Phase 6: Bubble Tea TUI Model
Build the interactive terminal UI model for the `add` command to ensure safe, user-friendly data entry without violating the schema. The dialogue collects the feature title, test type, implementation status, an optional description, an optional background section (multi-line via `textarea`), scenarios (multi-line, separated by blank lines), and examples (CSV format).

States 0–3 use `Enter` to advance. States 4–6 use `textarea` for multi-line input and `Ctrl+D` to advance.

```go
package main

import (
	"fmt"
	"strings"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type addModel struct {
	state           int // 0: Feature, 1: Type, 2: Status, 3: Description, 4: Background, 5: Scenarios, 6: Examples
	feature         textinput.Model
	description     textinput.Model
	background      textarea.Model
	scenarios       textarea.Model
	examples        textarea.Model
	testType        string
	status          string
	parsedScenarios []string
	parsedExamples  [][]string
	choices         []string
	cursor          int
	canceled        bool
	done            bool
}

func initialModel() addModel {
	fi := textinput.New(); fi.Placeholder = "Enter feature title"; fi.Focus()
	di := textinput.New(); di.Placeholder = "Enter description (optional)"
	bi := textarea.New(); bi.Placeholder = "Enter background / test environment setup (optional)"; bi.MaxHeight = 6; bi.ShowLineNumbers = false
	si := textarea.New(); si.Placeholder = "Enter Scenarios (separate each scenario with a blank line)"; si.MaxHeight = 8; si.ShowLineNumbers = false
	ei := textarea.New(); ei.Placeholder = "Enter Examples as CSV (header row first, then data rows)"; ei.MaxHeight = 8; ei.ShowLineNumbers = false
	return addModel{state: 0, feature: fi, description: di, background: bi, scenarios: si, examples: ei, parsedScenarios: []string{}, parsedExamples: [][]string{}}
}

func (m addModel) Init() tea.Cmd { return textinput.Blink }

func (m addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true; return m, tea.Quit
		case tea.KeyCtrlD: // Ctrl+D to finish textarea states
			switch m.state {
			case 4: m.state++; m.scenarios.Focus(); return m, textarea.Blink
			case 5: m.parsedScenarios = parseScenarios(m.scenarios.Value()); m.state++; m.examples.Focus(); return m, textarea.Blink
			case 6: m.parsedExamples = parseExamples(m.examples.Value()); m.done = true; return m, tea.Quit
			}
		case tea.KeyEnter:
			if m.state == 0 && m.feature.Value() != "" {
				m.state++; m.choices = []string{"functional", "solution", "performance", "reliability", "security"}; m.cursor = 0; return m, nil
			} else if m.state == 1 {
				m.testType = m.choices[m.cursor]; m.state++; m.choices = []string{"planned", "implemented", "deprecated"}; m.cursor = 0; return m, nil
			} else if m.state == 2 {
				m.status = m.choices[m.cursor]; m.state++; m.description.Focus(); return m, textinput.Blink
			} else if m.state == 3 {
				m.state++; m.background.Focus(); return m, textarea.Blink
			}
		case tea.KeyUp:
			if (m.state == 1 || m.state == 2) && m.cursor > 0 { m.cursor-- }
		case tea.KeyDown:
			if (m.state == 1 || m.state == 2) && m.cursor < len(m.choices)-1 { m.cursor++ }
		}
	}
	switch m.state {
	case 0: m.feature, cmd = m.feature.Update(msg)
	case 3: m.description, cmd = m.description.Update(msg)
	case 4: m.background, cmd = m.background.Update(msg)
	case 5: m.scenarios, cmd = m.scenarios.Update(msg)
	case 6: m.examples, cmd = m.examples.Update(msg)
	}
	return m, cmd
}

func (m addModel) View() string {
	if m.canceled { return "Operation canceled.\n" }
	if m.done { return "Saving test plan...\n" }
	s := "\n"
	switch m.state {
	case 0: s += "Feature Title:\n" + m.feature.View()
	case 1, 2:
		prompt := "Select Test Type:\n"
		if m.state == 2 { prompt = "Select Implementation Status:\n" }
		s += prompt
		for i, choice := range m.choices {
			cursor := " "; if m.cursor == i { cursor = ">" }
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	case 3: s += "Description (optional):\n" + m.description.View()
	case 4: s += "Background (optional) - Press Ctrl+D to finish:\n" + m.background.View()
	case 5: s += "Scenarios (Separate each scenario with a blank line) - Press Ctrl+D to finish:\n" + m.scenarios.View()
	case 6: s += "Examples (CSV format: header row first, then data rows) - Press Ctrl+D to finish:\n" + m.examples.View()
	}
	hint := "(Press Esc to quit)"
	if m.state >= 4 { hint += " | Ctrl+D to finish" }
	return s + "\n\n" + hint + "\n"
}

// parseScenarios splits multi-line text into individual scenarios separated by blank lines.
func parseScenarios(text string) []string {
	if text == "" { return []string{} }
	blocks := strings.Split(text, "\n\n")
	var scenarios []string
	for _, block := range blocks {
		trimmed := strings.TrimSpace(block)
		if trimmed != "" { scenarios = append(scenarios, trimmed) }
	}
	return scenarios
}

// parseExamples parses CSV-formatted text into a slice of string slices.
func parseExamples(text string) [][]string {
	if text == "" { return [][]string{} }
	lines := strings.Split(text, "\n")
	var result [][]string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" { continue }
		fields := strings.Split(line, ",")
		for i, field := range fields { fields[i] = strings.TrimSpace(field) }
		result = append(result, fields)
	}
	return result
}
```

---

### Phase 7: Cobra CLI Construction
Wire all features into a cohesive CLI with the four requested subcommands (`init`, `generate`, `add`, `serve`). The `serve` command utilizes `fsnotify` for live-reloading.

```go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "[github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)"
	"[github.com/fsnotify/fsnotify](https://github.com/fsnotify/fsnotify)"
	"[github.com/spf13/cobra](https://github.com/spf13/cobra)"
	"[github.com/spf13/viper](https://github.com/spf13/viper)"
	"go.yaml.in/yaml/v3"
)

var rootCmd = &cobra.Command{
	Use:   "gherkinator",
	Short: "A testing plan management and generation tool",
}

func main() {
	cobra.OnInitialize(initConfig)

	// 1. INIT COMMAND
	var initCmd = &cobra.Command{
		Use:   "init [directory-name]",
		Short: "Initialize a new test plan directory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dirName := args[0]
			os.MkdirAll(dirName, 0755)
			emptyPlan := "---\nfeature: \"\"\ntype: \"\"\nstatus: \"\"\ndescription: \"\"\nscenarios:\n  - \"\"\nexamples: []\n"
			filePath := filepath.Join(dirName, "test-plan.yaml")
			os.WriteFile(filePath, []byte(emptyPlan), 0644)
			fmt.Printf("Successfully initialized '%s'\n", filePath)
		},
	}

	// 2. GENERATE COMMAND
	var outputDir string
	var format string
	var generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate gherkin (gh) or markdown (md) files",
		Run: func(cmd *cobra.Command, args []string) {
			if format != "gh" && format != "md" {
				fmt.Println("Error: --format must be either 'gh' or 'md'")
				os.Exit(1)
			}
			if err := ProcessFile("test-plan.yaml", format, outputDir); err != nil {
				fmt.Printf("Generation failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Generation complete.")
		},
	}
	generateCmd.Flags().StringVarP(&outputDir, "output-dir", "o", ".", "Directory to save generated files")
	generateCmd.Flags().StringVar(&format, "format", "gh", "Output format (gh or md)")

	// 3. ADD COMMAND (Bubble Tea)
	var addCmd = &cobra.Command{
		Use:   "add",
		Short: "Interactively add a new test plan to test-plan.yaml",
		Run: func(cmd *cobra.Command, args []string) {
			m, err := tea.NewProgram(initialModel()).Run()
			if err != nil { os.Exit(1) }
			finalModel := m.(addModel)
			if finalModel.canceled { os.Exit(0) }

			desc := finalModel.description.Value()
			scenarios := finalModel.parsedScenarios
			if len(scenarios) == 0 {
				scenarios = []string{"Default Scenario\nGiven ...\nWhen ...\nThen ..."}
			}
			newPlan := TestPlan{
				Feature:     finalModel.feature.Value(),
				Type:        finalModel.testType,
				Status:      finalModel.status,
				Description: &desc,
				Scenarios:   scenarios,
			}
			bg := finalModel.background.Value()
			if bg != "" { newPlan.Background = &bg }
			if len(finalModel.parsedExamples) > 0 { newPlan.Examples = finalModel.parsedExamples }

			yamlData, _ := yaml.Marshal(&newPlan)
			f, _ := os.OpenFile("test-plan.yaml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			defer f.Close()
			f.WriteString("\n---\n" + string(yamlData))
			fmt.Println("Successfully appended new plan to test-plan.yaml")
		},
	}

	// 4. SERVE COMMAND
	var serveInputFile string
	var serveName string
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Serve test plan docs and watch for changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Derive project name: --name flag, or the directory name of the input file
			projectName := serveName
			if projectName == "" {
				absInput, _ := filepath.Abs(serveInputFile)
				projectName = filepath.Base(filepath.Dir(absInput))
			}

			tmpDir := "./.gherkindocs"
			os.RemoveAll(tmpDir)

			// Step 1: Clone slim Sphinx starter pack
			exec.Command(viper.GetString("tools.git"), "clone",
				"https://github.com/canonical/slim-sphinx-docs-starter-pack.git", tmpDir).Run()

			// Steps 2-4: Generate docs into type dirs, build toctree index, update conf.py
			if err := PrepareSphinxSite(serveInputFile, tmpDir, projectName); err != nil {
				return fmt.Errorf("failed to prepare sphinx site: %w", err)
			}

			docsDir := filepath.Join(tmpDir, "docs")

			// fsnotify Watcher for Live Reloading
			watcher, _ := fsnotify.NewWatcher()
			defer watcher.Close()
			go func() {
				for {
					select {
					case event, ok := <-watcher.Events:
						if !ok { return }
						if event.Has(fsnotify.Write) {
							fmt.Println("Change detected. Rebuilding docs...")
							PrepareSphinxSite(serveInputFile, tmpDir, projectName)
						}
					case <-watcher.Errors:
					}
				}
			}()
			watcher.Add(serveInputFile)

			// Step 5: Run make run inside a Bubbletea TUI for clean Ctrl+C handling
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

	rootCmd.AddCommand(initCmd, generateCmd, addCmd, serveCmd)
	if err := rootCmd.Execute(); err != nil { os.Exit(1) }
}
```

The `serve` command follows a five-step pipeline:

1. **Derive project name** — use the `--name` flag if provided; otherwise default to the base name of the directory containing the input YAML file (e.g. `charmed-hpc/test-plan.yaml` → `charmed-hpc`).
2. **Clone** the [slim Sphinx docs starter pack](https://github.com/canonical/slim-sphinx-docs-starter-pack) into `.gherkindocs`. This starter pack ships with just an `index.md`, `conf.py`, `Makefile`, and supporting Sphinx configuration — no extra content to prune.
3. **Generate** Markdown files using the same transpilation engine as `gherkinator generate --format md`, placing each file into its test-type subdirectory inside `docs/` (e.g. `.gherkindocs/docs/functional/login_feature.md`).
4. **Build the toctree and update conf.py** — create a sub-landing page (`index.md`) for each test type that has plans, then write the root `index.md` inside `docs/` with level-2 headers for each test type and bullet-pointed feature name links. A hidden `{toctree}` directive is also emitted so Sphinx still discovers all pages. The `UpdateConfPy` function then rewrites `conf.py` to: set the `project` variable to the derived project name, uncomment `disable_feedback_button = True` to hide the Sphinx feedback button, remove all `rediraffe`-related configuration lines (which cause build errors), and strip `"sphinx_rerediraffe"` from the extensions list.
5. **Serve** — run `make run` from inside the `.gherkindocs/docs` directory inside a **Bubbletea TUI** (`serveModel` in `serve_tui.go`). The TUI streams `make run` stdout/stderr line-by-line into a scrollable log view, keeps the last 100 lines, and handles **Ctrl+C** cleanly — killing the child process and exiting without leftover output polluting the terminal.

The `PrepareSphinxSite` function orchestrates steps 3–4, operating on the `docs/` subdirectory of the cloned repository:

```go
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

var validTestTypes = []string{
	"functional", "solution", "performance", "reliability", "security",
}

// GenerateSphinxDocs reads the YAML test plan file and writes Markdown
// files into <docsDir>/<type>/<safe_feature_name>.md.
func GenerateSphinxDocs(yamlFile string, docsDir string) ([]TestPlan, error) {
	file, _ := os.Open(yamlFile)
	defer file.Close()

	var plans []TestPlan
	decoder := yaml.NewDecoder(file)
	for i := 1; ; i++ {
		var plan TestPlan
		if err := decoder.Decode(&plan); err != nil {
			if err == io.EOF { break }
			return nil, err
		}
		if err := ValidateSchema(plan); err != nil { return nil, err }
		output := GenerateMarkdown(plan)
		safeFilename := strings.ReplaceAll(strings.ToLower(plan.Feature), " ", "_")
		if safeFilename == "" { safeFilename = fmt.Sprintf("plan_%d", i) }
		typeDir := filepath.Join(docsDir, plan.Type)
		os.MkdirAll(typeDir, 0755)
		os.WriteFile(filepath.Join(typeDir, safeFilename+".md"), []byte(output), 0644)
		plans = append(plans, plan)
	}
	return plans, nil
}

// BuildTypeLandingPages creates a sub-landing page (index.md) for each
// test type that has plans, with a toctree listing the plan files.
func BuildTypeLandingPages(docsDir string, grouped map[string][]string) error {
	for _, testType := range validTestTypes {
		files := grouped[testType]
		if len(files) == 0 { continue }
		var builder strings.Builder
		fmt.Fprintf(&builder, "# %s\n\n```{toctree}\n:maxdepth: 1\n\n", strings.Title(testType))
		for _, f := range files { fmt.Fprintf(&builder, "%s\n", f) }
		builder.WriteString("```\n")
		typeDir := filepath.Join(docsDir, testType)
		os.MkdirAll(typeDir, 0755)
		os.WriteFile(filepath.Join(typeDir, "index.md"), []byte(builder.String()), 0644)
	}
	return nil
}

// planEntry stores a feature name alongside its safe filename so that
// the root index can list human-readable bullet points.
type planEntry struct {
	Feature      string
	SafeFilename string
}

// BuildSphinxIndex writes the root index.md with level-2 headers for each
// test type and bullet-pointed feature name links underneath. A hidden
// toctree is emitted so Sphinx still discovers all pages.
func BuildSphinxIndex(docsDir string, plans []TestPlan) error {
	grouped := make(map[string][]string)
	entries := make(map[string][]planEntry)
	for _, plan := range plans {
		safe := strings.ReplaceAll(strings.ToLower(plan.Feature), " ", "_")
		if safe == "" { safe = "plan" }
		grouped[plan.Type] = append(grouped[plan.Type], safe)
		entries[plan.Type] = append(entries[plan.Type], planEntry{
			Feature: plan.Feature, SafeFilename: safe,
		})
	}
	BuildTypeLandingPages(docsDir, grouped)

	var builder strings.Builder
	builder.WriteString("# Test Plans\n\n")

	// Hidden toctree so Sphinx still discovers all pages
	builder.WriteString("```{toctree}\n:hidden:\n:maxdepth: 2\n\n")
	for _, testType := range validTestTypes {
		if len(grouped[testType]) == 0 { continue }
		fmt.Fprintf(&builder, "%s/index\n", testType)
	}
	builder.WriteString("```\n\n")

	// Visible type sections with feature bullet points
	for _, testType := range validTestTypes {
		typeEntries := entries[testType]
		if len(typeEntries) == 0 { continue }
		fmt.Fprintf(&builder, "## %s\n\n", strings.Title(testType))
		for _, entry := range typeEntries {
			name := entry.Feature
			if name == "" { name = entry.SafeFilename }
			fmt.Fprintf(&builder, "- [%s](%s/%s.md)\n", name, testType, entry.SafeFilename)
		}
		builder.WriteString("\n")
	}

	return os.WriteFile(filepath.Join(docsDir, "index.md"), []byte(builder.String()), 0644)
}

// UpdateConfPy rewrites the Sphinx conf.py file to:
//   - set the `project` variable to the given name
//   - uncomment `disable_feedback_button = True`
//   - remove rediraffe-related configuration lines
//   - remove "sphinx_rerediraffe" from the extensions list
func UpdateConfPy(confPath string, projectName string) error {
	data, _ := os.ReadFile(confPath)
	lines := strings.Split(string(data), "\n")
	var out []string
	foundProject := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !foundProject && strings.HasPrefix(trimmed, "project") && strings.Contains(trimmed, "=") {
			out = append(out, fmt.Sprintf("project = %q", projectName))
			foundProject = true
			continue
		}
		if trimmed == "# disable_feedback_button = True" {
			out = append(out, "disable_feedback_button = True")
			continue
		}
		if strings.HasPrefix(trimmed, "rediraffe_redirects") ||
			strings.HasPrefix(trimmed, "rediraffe_dir_only") {
			continue
		}
		if strings.Contains(trimmed, `"sphinx_rerediraffe"`) {
			continue
		}
		out = append(out, line)
	}
	if !foundProject {
		return fmt.Errorf("could not find 'project' variable in %s", confPath)
	}
	return os.WriteFile(confPath, []byte(strings.Join(out, "\n")), 0644)
}

// PrepareSphinxSite orchestrates steps 3–4, operating on <cloneDir>/docs/.
func PrepareSphinxSite(yamlFile string, cloneDir string, projectName string) error {
	docsDir := filepath.Join(cloneDir, "docs")
	plans, err := GenerateSphinxDocs(yamlFile, docsDir)
	if err != nil { return err }
	if err := BuildSphinxIndex(docsDir, plans); err != nil { return err }
	return UpdateConfPy(filepath.Join(docsDir, "conf.py"), projectName)
}
```

---

### Phase 8: Delete Command
Implement a `delete` subcommand that removes test plans from `test-plan.yaml` by feature name. The command accepts multiple feature names, performs case-insensitive matching, and prompts for confirmation unless the `-y/--yes` flag is provided.

```go
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"go.yaml.in/yaml/v3"
)

// LoadTestPlans reads a multi-document YAML file and returns all TestPlan entries.
func LoadTestPlans(filename string) ([]TestPlan, error) {
	file, err := os.Open(filename)
	if err != nil { return nil, fmt.Errorf("failed to open file: %w", err) }
	defer file.Close()

	var plans []TestPlan
	decoder := yaml.NewDecoder(file)
	for {
		var plan TestPlan
		if err := decoder.Decode(&plan); err != nil {
			if err == io.EOF { break }
			return nil, fmt.Errorf("failed to decode YAML: %w", err)
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

// WriteTestPlans writes a slice of TestPlan entries to a YAML file as a
// multi-document stream separated by "---".
func WriteTestPlans(filename string, plans []TestPlan) error {
	file, err := os.Create(filename)
	if err != nil { return fmt.Errorf("failed to create file: %w", err) }
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	for _, plan := range plans {
		if err := encoder.Encode(&plan); err != nil { return fmt.Errorf("failed to encode YAML: %w", err) }
	}
	return nil
}

// DeleteTestPlans removes test plans whose feature names match the provided
// list (case-insensitive). It returns the remaining plans and the names that
// were actually found and deleted.
func DeleteTestPlans(plans []TestPlan, featureNames []string) (remaining []TestPlan, deleted []string) {
	toDelete := make(map[string]bool)
	for _, name := range featureNames { toDelete[strings.ToLower(name)] = true }

	for _, plan := range plans {
		if toDelete[strings.ToLower(plan.Feature)] {
			deleted = append(deleted, plan.Feature)
		} else {
			remaining = append(remaining, plan)
		}
	}
	return remaining, deleted
}

// ConfirmDeletion prompts the user for confirmation and reads from the
// provided reader. Returns true only when the user enters "Y".
func ConfirmDeletion(featureNames []string, reader io.Reader) bool {
	fmt.Fprintf(os.Stdout, "Are you sure you want to delete test plans %s? [Y/n] ",
		strings.Join(quoteNames(featureNames), ", "))

	scanner := bufio.NewScanner(reader)
	if scanner.Scan() { return strings.TrimSpace(scanner.Text()) == "Y" }
	return false
}
```

The delete command is wired into the Cobra CLI as follows:

```go
// DELETE COMMAND
var skipConfirm bool
var deleteInputFile string
var deleteCmd = &cobra.Command{
	Use:   "delete [feature-names...]",
	Short: "Delete test plans by feature name from test-plan.yaml",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		plans, err := LoadTestPlans(deleteInputFile)
		if err != nil { return fmt.Errorf("failed to load test plans: %w", err) }

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
```

**Usage examples:**
```bash
# Delete with confirmation prompt
gherkinator delete "job submission" "gpu job submission"

# Delete without confirmation
gherkinator delete -y "job submission" "gpu job submission"

# Delete from a specific file
gherkinator delete -y -i my-plans.yaml "job submission"
```

---

### Phase 9: Clean Command
Implement a `clean` subcommand that removes generated files (`.feature`, `.md`) and the `.gherkindocs` hidden directory from the test plan directory. This helps keep the working directory tidy after generation or serving.

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CleanDirectory removes generated files (.feature, .md) and the
// .gherkindocs hidden directory from the specified directory.
func CleanDirectory(dir string) error {
	// Remove the hidden docs serve directory
	serveDir := filepath.Join(dir, ".gherkindocs")
	if err := os.RemoveAll(serveDir); err != nil {
		return fmt.Errorf("failed to remove %s: %w", serveDir, err)
	}

	// Remove generated .feature and .md files
	entries, err := os.ReadDir(dir)
	if err != nil { return fmt.Errorf("failed to read directory %s: %w", dir, err) }

	for _, entry := range entries {
		if entry.IsDir() { continue }
		name := entry.Name()
		lower := strings.ToLower(name)
		if strings.HasSuffix(lower, ".feature") || strings.HasSuffix(lower, ".md") {
			path := filepath.Join(dir, name)
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", path, err)
			}
			fmt.Printf("Removed %s\n", path)
		}
	}
	return nil
}
```

The clean command is wired into the Cobra CLI as follows:

```go
// CLEAN COMMAND
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
```

**Usage examples:**
```bash
# Clean the current directory
gherkinator clean

# Clean a specific directory
gherkinator clean -d ./my-test-plans
```

---

### Phase 10: Unit Testing with Testify (100% Coverage)
Ensure a robust foundation using `github.com/stretchr/testify/assert` and `require`.

```go
package main

import (
	"testing"
	"[github.com/stretchr/testify/assert](https://github.com/stretchr/testify/assert)"
)

func TestValidateSchema(t *testing.T) {
	tests := []struct {
		name        string
		plan        TestPlan
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid Plan",
			plan: TestPlan{Type: "functional", Status: "planned"},
			expectError: false,
		},
		{
			name: "Invalid Type",
			plan: TestPlan{Type: "invalid_type", Status: "implemented"},
			expectError: true,
			errorMsg: "invalid type 'invalid_type'",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSchema(tc.plan)
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

---

### Phase 11: Task Management with `just`
Streamline development workflows using the `just` command runner. The `justfile` includes targets for formatting, linting, testing, packaging, and cleaning.

```justfile
# justfile
set shell := ["bash", "-c"]

# The default recipe runs when you type `just`
default: build

# Build the gherkinator binary
build: fmt
    @echo "Building gherkinator..."
    go build -o gherkinator .

# Apply formatting standards to project
fmt:
    @echo "Formatting Go code..."
    go fmt ./...

# Check project against coding style standards
lint:
    @echo "Running linter..."
    golangci-lint run ./...

# Run unit tests with testify and generate coverage profile
unit:
    @echo "Running unit tests..."
    go test -v -coverprofile=coverage.out ./...

# View HTML coverage report
coverage: unit
    @echo "Generating coverage report..."
    go tool cover -html=coverage.out

# Install the binary to the system GOPATH
install: build
    @echo "Installing gherkinator..."
    go install .

# Build the snap package using snapcraft
snap:
    @echo "Building snap package..."
    snapcraft

# Clean build artifacts, temporary doc server files, and snapcraft artifacts
clean:
    @echo "Cleaning up workspace..."
    rm -f gherkinator coverage.out
    rm -rf .gherkindocs
    @echo "Cleaning snapcraft cache and artifacts..."
    snapcraft clean || true
    rm -f *.snap
```

---

### Phase 12: Snap Packaging with Snapcraft
Package `gherkinator` for easy distribution across Linux systems utilizing a `snapcraft.yaml` configuration with `classic` confinement.

```yaml
# snap/snapcraft.yaml
name: gherkinator
base: core24
version: 0.1.0
summary: A testing plan management and generation tool
description: |
  Gherkinator parses, validates, and transpiles centralized YAML test plans 
  into Gherkin feature files or Markdown documentation. It also provides an 
  interactive local documentation server using the Canonical Sphinx starter pack.

grade: stable
confinement: classic 

apps:
  gherkinator:
    command: bin/gherkinator

parts:
  gherkinator:
    plugin: go
    source: .
    build-snaps:
      - go
```
