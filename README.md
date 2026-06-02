# gherkinator

A CLI tool for managing centralised YAML test plans. It validates, transpiles,
and serves test plans as Gherkin feature files or Sphinx-powered Markdown
documentation.

## Overview

`gherkinator` treats test plans as structured YAML documents. From a single
source of truth you can:

- **Generate** Gherkin `.feature` files ready for BDD frameworks.
- **Generate** Markdown documentation with bolded step keywords.
- **Serve** a live Sphinx documentation site that rebuilds whenever the YAML
  changes.

Multiple test plans can live in one YAML file, separated by `---`.

## Prerequisites

| Tool | Required for |
| --- | --- |
| Go 1.26+ | building from source |
| `git` | `serve` (clones the Sphinx starter pack) |
| `python3`, `pip` | `serve` (Sphinx build environment) |
| `make` | `serve` (runs `make run` inside the docs directory) |

## Installation

### Snap (recommended)

```bash
sudo snap install gherkinator --classic
```

### Build from source

```bash
git clone https://github.com/canonical/gherkinator.git
cd gherkinator
go build -o gherkinator .
```

## YAML schema

Each document in a `test-plan.yaml` file maps to the following schema.

```yaml
# ── Required fields ──────────────────────────────────────────────────────────

feature: "GPU job submission"   # Human-readable feature name

type: functional                # One of: functional | solution | performance
                                #         reliability | security

status: planned                 # One of: planned | implemented | deprecated

risk: stable                    # One of: edge | beta | candidate | stable

scenarios:                      # At least one scenario string.
  - |-                          # The first line is the scenario title;
    Default Scenario            # subsequent lines are Gherkin steps.
    Given the system is running
    When a job is submitted
    Then the job completes successfully

# ── Optional fields ──────────────────────────────────────────────────────────

description: "Test GPU job submission on Charmed HPC"

issues: "https://github.com/canonical/charmed-hpc/issues/42"

docs: "https://docs.canonical.com/charmed-hpc"

background: |-                  # Steps that run before every scenario.
  Given the cluster is available
  And I am logged in as a user

examples:                       # Parametrised data rows.  Use <param> tokens
  - - alice                     # in scenario steps; the first row supplies
    - admin                     # column headers derived from those tokens.
  - - bob
    - viewer
```

### Parametrised scenarios

When `examples` is present and scenarios reference `<param>` placeholders:

- The Gherkin output uses `Scenario Outline:` and an `Examples:` table.
- The Markdown output uses a `## Scenario Outlines` section and a
  `## Examples` table whose header row is extracted from the `<param>`
  tokens in the scenario text.

### Multi-document files

A single `test-plan.yaml` can contain multiple plans separated by `---`:

```yaml
feature: "Login Feature"
type: functional
status: planned
scenarios:
  - |-
    User logs in
    Given a user exists
    When the user enters valid credentials
    Then the user sees the dashboard
---
feature: "Stress Test"
type: performance
status: implemented
scenarios:
  - |-
    Load test
    Given the system is running
    When 1000 users connect simultaneously
    Then response time is under 500ms
```

## Commands

### `init`

Initialise a new test plan directory with an empty `test-plan.yaml`.

```
gherkinator init <directory-name>
```

**Example:**

```bash
gherkinator init charmed-hpc
# Creates charmed-hpc/test-plan.yaml
```

---

### `generate`

Transpile a YAML test plan into Gherkin feature files or Markdown documents.

```
gherkinator generate [flags]
```

| Flag | Short | Default | Description |
| --- | --- | --- | --- |
| `--format` | | `gh` | Output format: `gh` (Gherkin) or `md` (Markdown) |
| `--input` | `-i` | `test-plan.yaml` | Path to the input YAML file |
| `--output-dir` | `-o` | `.` | Directory to write output files into |
| `--risk` | | | Filter by risk level: `edge`, `beta`, `candidate`, `stable` (cumulative) |

**Examples:**

```bash
# Generate .feature files
gherkinator generate --format gh -i charmed-hpc/test-plan.yaml -o charmed-hpc

# Generate Markdown files
gherkinator generate --format md -i charmed-hpc/test-plan.yaml -o charmed-hpc

# Generate only edge risk plans
gherkinator generate --format md -i charmed-hpc/test-plan.yaml -o charmed-hpc --risk edge

# Generate edge and beta risk plans
gherkinator generate --format md -i charmed-hpc/test-plan.yaml -o charmed-hpc --risk beta
```

Output filenames are derived from the `feature` field
(`"GPU job submission"` → `gpu_job_submission.feature` / `.md`).

---

### `serve`

Serve the test plans as a live Sphinx documentation site.

```
gherkinator serve [flags]
```

| Flag | Short | Default | Description |
| --- | --- | --- | --- |
| `--input` | `-i` | `test-plan.yaml` | Path to the input YAML file |
| `--name` | `-n` | input directory name | Project name shown in the docs |
| `--risk` | | | Filter by risk level: `edge`, `beta`, `candidate`, `stable` (cumulative) |

The command follows this pipeline:

1. Derives the project name from `--name`, or falls back to the base name of
   the directory containing the input file
   (e.g. `charmed-hpc/test-plan.yaml` → `charmed-hpc`).
2. Clones the
   [Canonical Slim Sphinx docs starter pack](https://github.com/canonical/slim-sphinx-docs-starter-pack)
   into `.gherkindocs/`.
3. Generates Markdown files into type-based subdirectories inside
   `.gherkindocs/docs/`.
4. Builds a root `index.md` with level-2 headers per test type and
   bullet-pointed feature links; patches `conf.py` to set the project name,
   disable the feedback button, and remove `rediraffe` configuration that
   causes build errors.
5. Runs `make run` inside a Bubbletea TUI that streams build/server logs.
   Press **Ctrl+C** to stop the server cleanly.

The YAML file is watched for changes; the docs rebuild automatically on save.

**Examples:**

```bash
# Serve with default project name (derived from directory)
gherkinator serve -i charmed-hpc/test-plan.yaml

# Override the project name shown in the docs
gherkinator serve -i charmed-hpc/test-plan.yaml --name "Charmed HPC"

# Serve only edge and beta risk plans
gherkinator serve -i charmed-hpc/test-plan.yaml --risk beta
```

---

### `delete`

Remove one or more test plans from `test-plan.yaml` by feature name
(case-insensitive).

```
gherkinator delete [feature-names...] [flags]
```

| Flag | Short | Default | Description |
| --- | --- | --- | --- |
| `--yes` | `-y` | `false` | Skip the confirmation prompt |
| `--input` | `-i` | `test-plan.yaml` | Path to the input YAML file |

**Examples:**

```bash
# Interactive confirmation
gherkinator delete "GPU job submission"

# Delete multiple plans without confirmation
gherkinator delete -y "GPU job submission" "MPI job submission"

# Delete from a specific file
gherkinator delete -y -i charmed-hpc/test-plan.yaml "GPU job submission"
```

---

### `clean`

Remove generated files (`.feature`, `.md`) and the `.gherkindocs` temporary
directory from a test plan directory.

```
gherkinator clean [flags]
```

| Flag | Short | Default | Description |
| --- | --- | --- | --- |
| `--dir` | `-d` | `.` | Directory to clean |

**Examples:**

```bash
# Clean the current directory
gherkinator clean

# Clean a specific directory
gherkinator clean -d charmed-hpc
```

## Configuration

Tool paths used by the `serve` command can be overridden via a
`.gherkinator.yaml` file (searched in `.` then `$HOME`) or environment
variables prefixed with `GHERKINATOR_`.

**`.gherkinator.yaml`:**

```yaml
tools:
  git: /usr/bin/git
  python3: /usr/bin/python3
  pip: /usr/bin/pip3
  make: /usr/bin/make
```

**Environment variables:**

| Variable | Default |
| --- | --- |
| `GHERKINATOR_TOOLS_GIT` | `git` |
| `GHERKINATOR_TOOLS_PYTHON3` | `python3` |
| `GHERKINATOR_TOOLS_PIP` | `pip` |
| `GHERKINATOR_TOOLS_MAKE` | `make` |

