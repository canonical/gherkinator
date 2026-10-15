---
name: gherkinator-cli
description: Use the gherkinator CLI to manage centralized YAML test plans. Covers init, edit, validate, generate, serve, delete, clean, and diff commands, the YAML test plan schema (including custom classification tags), risk/status/tag filtering, and CI drift checking. Use when creating, editing, validating, transpiling, or serving Gherkin feature files or Markdown docs from YAML test plans, when filtering plans by risk, status, or deployment tags, or when checking that committed .feature files match their source YAML.
---

# gherkinator CLI

`gherkinator` treats YAML test plans as a single source of truth and
transpiles them into Gherkin `.feature` files, Markdown documentation, or a
live Sphinx docs site.

## When to use this skill

- Creating or scaffolding a new test plan directory (`init`).
- Writing or editing YAML test plans (`edit`, or edit the YAML directly).
- Validating test plans against the schema (`validate`).
- Generating `.feature` or `.md` output (`generate`).
- Serving live docs with rebuild-on-save (`serve`).
- Removing test plans or generated artifacts (`delete`, `clean`).
- Checking committed `.feature` files match source YAML in CI (`diff`).
- Filtering plans by risk level, implementation status, or deployment tags.

## Core workflow

1. `gherkinator init <dir>` to scaffold (`-n` names the YAML file), or edit
   YAML directly.
2. Author plans following the schema in `reference.md` (required fields:
   `feature`, `type`, `status`, `risk`, `scenarios`).
3. `gherkinator validate <files-or-dirs>` — always validate before
   generating.
4. `gherkinator generate --format gh|md <inputs> -o <outdir>` to produce
   output, optionally filtered with `--risk` / `--status` / `--tag`.
5. `gherkinator diff <plans> <features>` in CI to catch drift between YAML
   and committed `.feature` files.

## Command quick reference

| Command | Purpose |
| --- | --- |
| `init [dir]` | Create a directory with an empty test plan template (`-n` names the YAML file) |
| `edit <file>` | Open the plan in `$VISUAL`/`$EDITOR` with schema help; validates on save |
| `validate [inputs]` | Check plans against the schema |
| `generate [inputs]` | Emit `.feature` (`--format gh`) or `.md` (`--format md`) files |
| `serve [inputs]` | Live Sphinx docs site; rebuilds on YAML save |
| `delete [names...]` | Remove plans by feature name (requires `-i <file>`) |
| `clean` | Remove generated `.feature`/`.md` files and `.gherkindocs/` (`-d` selects the directory) |
| `diff <plans> <features>` | CI drift check; exit 0 = in sync, 1 = diffs, 2 = error |
| `version` / `--version` | Print version (note: `-v` is NOT a version shorthand) |

Input arguments for `validate`, `generate`, and `serve` accept any mix of
`.yaml`/`.yml` files and directories (scanned non-recursively). With no
arguments, the current working directory is scanned.

## Key rules

- **Schema enums**: `type` ∈ {functional, solution, performance,
  reliability, security}; `status` ∈ {planned, implemented, deprecated};
  `risk` ∈ {edge, beta, candidate, stable}.
- **Risk filtering is cumulative**: `--risk beta` includes edge + beta.
- **Custom tags**: plans may carry an optional `tags` list of freeform
  classification tags (e.g. `single-node`, `multi-node`) for filtering by
  deployment configuration. Each tag must be non-empty with no whitespace
  or commas.
- **Tag filtering is union-based**: `--tag a --tag b` (or `--tag a,b`)
  renders plans carrying **at least one** requested tag. Matching is
  exact and case-sensitive; untagged plans are excluded while the filter
  is set. `--tag` is intersected with `--risk` and `--status` (a plan
  must satisfy every set filter dimension).
- **Rendered tag line**: `.feature` output starts with a feature-level
  tag line `@<type> @<risk> @<status> [@<custom-tag>...]` directly above
  the `Feature:` header; every scenario in the file inherits these tags.
  Markdown output renders `**Type:**`, `**Status:**`, `**Risk:**`, and
  (when present) `**Tags:**` bullets.
- **Changing `type`, `risk`, `status`, or `tags` changes rendered output**:
  regenerate and re-commit any `.feature` files checked into the repo —
  `diff` will report drift otherwise.
- **Multi-document files**: separate plans in one file with `---`; each
  document is validated independently.
- **Parametrized scenarios**: use `<param>` tokens in steps plus an
  `examples` table to get `Scenario Outline:` / `Examples:` output.
- **Output filenames** derive from the `feature` field
  (`"GPU job submission"` → `gpu_job_submission.feature`); an empty
  feature name falls back to `plan_N`.
- **`serve` prerequisites**: `git`, `python3`, `pip`, `make`. It clones the
  Canonical Sphinx stack into `.gherkindocs/` and runs `make run` in a TUI;
  stop with Ctrl+C.
- **`delete` requires `-i <file>`** and matches feature names
  case-insensitively; use `-y` to skip confirmation.
- **Never install the snap on the host or publish it** when working in the
  gherkinator repo itself.

## Detailed references

Read these files for full details when needed:

- `reference.md` — complete YAML schema (including tag format rules),
  rendered-output contract, all flags per command, config file and
  environment variables, exit codes.
- `examples.md` — end-to-end worked examples: authoring, validating,
  generating, tag/risk/status filtering, serving, deleting, and a CI
  diff check.
