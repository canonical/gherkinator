# gherkinator reference

Complete reference for the `gherkinator` CLI, verified against the source
in `cmd/gherkinator/cmd/` of github.com/canonical/gherkinator.

## YAML test plan schema

Each YAML document maps to one test plan. Multiple documents may live in a
single file, separated by `---`.

### Required fields

| Field | Type | Allowed values |
| --- | --- | --- |
| `feature` | string | Human-readable feature name; also derives the output filename |
| `type` | string | `functional`, `solution`, `performance`, `reliability`, `security` |
| `status` | string | `planned`, `implemented`, `deprecated` |
| `risk` | string | `edge`, `beta`, `candidate`, `stable` |
| `scenarios` | []string | At least one scenario. First line = scenario title; subsequent lines = Gherkin steps |

### Optional fields

| Field | Type | Meaning |
| --- | --- | --- |
| `description` | string | Free-form description |
| `issues` | string | Link to tracked issues |
| `docs` | string | Link to documentation |
| `background` | string | Gherkin steps that run before every scenario |
| `examples` | [][]string | Parametrized data rows; use `<param>` tokens in scenario steps. The first row supplies column headers derived from those tokens |
| `tags` | []string | Freeform classification tags (e.g. `single-node`, `multi-node`) used to filter plans by deployment configuration. Rendered as Gherkin tags and a Markdown bullet; filterable via `--tag` |

### Tag format rules

- Each tag must be non-empty and contain **no whitespace** (it would break
  Gherkin tag rendering) and **no commas** (the `--tag` flag's
  comma-separated list separator).
- Tags are validated both when plans are written (`validate`, `edit`) and
  when `--tag` filters are passed to `generate`, `serve`, and `diff`.

### Parametrized scenarios

When `examples` is present and scenarios reference `<param>` placeholders:

- Gherkin output uses `Scenario Outline:` and an `Examples:` table.
- Markdown output uses a `## Scenario Outlines` section and an
  `## Examples` table whose header row is extracted from the `<param>`
  tokens in the scenario text.

## Rendered output

Every plan renders its classification fields into the generated files:

- **Gherkin (`.feature`)**: a single feature-level tag line directly
  above the `Feature:` header, in order
  `@<type> @<risk> @<status> [@<custom-tag>...]`, for example
  `@functional @edge @planned @multi-node @minimal`. Feature-level tags
  are inherited by every scenario in the file, so downstream BDD runners
  can filter (or skip planned/deprecated scenarios) via tag expressions.
- **Markdown (`.md`)**: `**Type:**`, `**Status:**`, and `**Risk:**` bullets,
  plus a `**Tags:**` bullet (comma-separated) when the plan carries tags.

Because these fields are rendered, changing `type`, `risk`, `status`, or
`tags` changes the generated content. Regenerate and re-commit any
`.feature` files checked into the repository (see `diff`), or CI drift
checks will fail.

## Commands

### `init [directory-name]`

Creates the directory and an empty test plan template (including empty
`examples` and `tags` lists).

| Flag | Short | Default | Description |
| --- | --- | --- | --- |
| `--name` | `-n` | `test-plan.yaml` | YAML filename; `.yaml` appended unless a `.yaml`/`.yml` extension is already present |

Exactly one positional argument (the directory) is required.

### `edit <file>`

Opens the file in the user's editor (`$VISUAL`/`$EDITOR`) pre-populated
with a schema help comment block. On save, the buffer is parsed and
validated; only valid content is written back to the file. Exactly one
positional argument.

### `validate [files-or-directories...]`

No flags. On success prints `All N test plan(s) are valid.` On failure,
each invalid document is reported with source file, document index
(1-based), and reason, e.g.:

```
test plan 'test-plan.yaml' failed validation (document 1): invalid type 'bogus': must be one of 'functional', 'solution', 'performance', 'reliability', or 'security'
```

### `generate [files-or-directories...]`

| Flag | Short | Default | Description |
| --- | --- | --- | --- |
| `--format` | | `gh` | `gh` (Gherkin `.feature`) or `md` (Markdown) |
| `--output-dir` | `-o` | `.` | Directory to write output files into |
| `--risk` | | _(unset)_ | Cumulative filter: `edge`, `beta`, `candidate`, `stable` |
| `--status` | | _(unset)_ | Exact-match filter: `planned`, `implemented`, `deprecated` |
| `--tag` | | _(unset)_ | Union tag filter; repeatable (`--tag a --tag b`) or comma-separated (`--tag a,b`) |

Risk filtering is cumulative — `--risk beta` includes plans at the edge
and beta levels. Tag filtering uses union semantics — a plan is rendered
when it carries **at least one** requested tag; untagged plans are
excluded while the filter is set, and matching is exact and case-sensitive.
All set filters are intersected: a plan must satisfy every non-empty
filter dimension.

Output filenames derive from the `feature` field
(`"GPU job submission"` → `gpu_job_submission.feature` / `.md`); an empty
feature name falls back to `plan_N` (N = 1-based index of the rendered
plan).

### `serve [files-or-directories...]`

| Flag | Short | Default | Description |
| --- | --- | --- | --- |
| `--name` | `-n` | CWD base name | Project name shown in the docs |
| `--risk` | | _(unset)_ | Cumulative risk filter (same semantics as generate) |
| `--status` | | _(unset)_ | Status filter (same semantics as generate) |
| `--tag` | | _(unset)_ | Union tag filter (same semantics as generate) |

Pipeline: derives project name → clones the Canonical Sphinx stack
(github.com/canonical/sphinx-stack) into `.gherkindocs/` → prunes template
content dirs and default `index.rst` → generates Markdown into type-based
subdirectories under `.gherkindocs/docs/` → builds a root `index.md` and
patches `conf.py` → runs `make run` inside a Bubbletea TUI streaming
build/server logs. Input YAML files are watched; docs rebuild on save.
Stop with Ctrl+C.

Prerequisites: `git`, `python3`, `pip`, `make`.

### `delete [feature-names...]`

| Flag | Short | Description |
| --- | --- | --- |
| `--yes` | `-y` | Skip the confirmation prompt |
| `--input` | `-i` | Path to the input YAML file (**required**) |

Requires at least one feature name. Matching is case-insensitive. If no
plans match, prints `No matching test plans found.` and exits 0 without
modifying the file.

### `clean`

| Flag | Short | Default | Description |
| --- | --- | --- | --- |
| `--dir` | `-d` | `.` | Directory to clean |

Removes generated `.feature` and `.md` files plus the `.gherkindocs/`
staging directory.

### `diff [flags] <plans> <features>`

Exactly two positional arguments: `<plans>` (YAML file or directory, same
discovery rules as generate) and `<features>` (directory containing
on-disk `.feature` files).

| Flag | Default | Description |
| --- | --- | --- |
| `--risk` | _(unset)_ | Restrict comparison to plans at or below the given risk level |
| `--status` | _(unset)_ | Restrict comparison to plans with the given status |
| `--tag` | _(unset)_ | Restrict comparison to plans carrying at least one requested tag (union); untagged plans are excluded while set |
| `--show` | `false` | Print unified diffs (`diff -u` format) to stdout for every differing file |

Exit codes:

| Code | Meaning |
| --- | --- |
| `0` | No diffs — plans and feature files are in sync |
| `1` | At least one classification reported (missing, differing, or orphan) |
| `2` | Plan or features-directory error (bad schema, malformed YAML, missing dir, or two plans rendering to the same filename) |

Classifications (reported on stderr):

| Category | Triggered when |
| --- | --- |
| **Missing** | A plan has no corresponding `.feature` file under `<features>` |
| **Differing** | A plan's rendered `.feature` content does not match the file on disk |
| **Orphan** | A `.feature` file exists under `<features>` but no plan maps to it |

A feature file that maps to a plan but whose content differs is classified
as **differing**, not orphan.

### `version`

Prints the bare version string (e.g. `0.1.1`). Also available as the
top-level `--version` flag. The shorthand `-v` is intentionally NOT bound
to `--version` (reserved for a future `--verbose`).

## Input discovery rules

Shared by `validate`, `generate`, `serve`, and the `<plans>` argument of
`diff`:

- Any mix of `.yaml`/`.yml` files and directories.
- Directories are scanned **non-recursively** for YAML files.
- With no arguments, the current working directory is scanned.

## Configuration

Tool paths used by `serve` can be overridden via a `.gherkinator.yaml`
file (searched in `.` then `$HOME`) or environment variables prefixed with
`GHERKINATOR_`.

`.gherkinator.yaml`:

```yaml
tools:
  git: /usr/bin/git
  python3: /usr/bin/python3
  pip: /usr/bin/pip3
  make: /usr/bin/make
```

Environment variables:

| Variable | Default |
| --- | --- |
| `GHERKINATOR_TOOLS_GIT` | `git` |
| `GHERKINATOR_TOOLS_PYTHON3` | `python3` |
| `GHERKINATOR_TOOLS_PIP` | `pip` |
| `GHERKINATOR_TOOLS_MAKE` | `make` |
