# gherkinator examples

End-to-end worked examples for the gherkinator CLI.

## 1. Scaffold a new test plan directory

```bash
# Creates charmed-hpc/test-plan.yaml with an empty template
gherkinator init charmed-hpc

# Custom filename (extension auto-appended)
gherkinator init charmed-hpc --name my-test-plan
# Creates charmed-hpc/my-test-plan.yaml
```

## 2. Author a test plan

Write `charmed-hpc/test-plan.yaml`:

```yaml
feature: "GPU job submission"
type: functional
status: planned
risk: stable
description: "Test GPU job submission on Charmed HPC"
issues: "https://github.com/canonical/charmed-hpc/issues/42"
docs: "https://docs.canonical.com/charmed-hpc"
background: |-
  Given the cluster is available
  And I am logged in as a user
scenarios:
  - |-
    Default Scenario
    Given the system is running
    When a job is submitted
    Then the job completes successfully
tags:
  - single-node
```

Multiple plans in one file, separated by `---`:

```yaml
feature: "Login Feature"
type: functional
status: planned
risk: edge
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
risk: beta
tags:
  - multi-node
scenarios:
  - |-
    Load test
    Given the system is running
    When 1000 users connect simultaneously
    Then response time is under 500ms
```

Parametrized scenario using `<param>` tokens and an `examples` table:

```yaml
feature: "Role-based access"
type: security
status: implemented
risk: candidate
scenarios:
  - |-
    Access by role
    Given a user <name> with role <role>
    When the user requests the admin panel
    Then access is <outcome>
examples:
  - - alice
    - admin
    - granted
  - - bob
    - viewer
    - denied
```

Tag rules: each tag must be non-empty with no whitespace or commas
(e.g. `single-node`, `multi-node`, `minimal`). Tags classify plans by
deployment configuration and are inherited by every scenario in the
rendered feature file.

## 3. Edit a plan with schema help

```bash
gherkinator edit charmed-hpc/test-plan.yaml
```

Opens `$VISUAL`/`$EDITOR` with a schema reference comment block prepended.
On save, the content is validated; only valid plans are written back.

## 4. Validate

```bash
gherkinator validate charmed-hpc/test-plan.yaml
# All 1 test plan(s) are valid.

gherkinator validate charmed-hpc/plans/          # whole directory
gherkinator validate plans/alpha.yaml plans/beta.yml
gherkinator validate                             # scan CWD
```

Failure output names the file, document index, and reason:

```
test plan 'charmed-hpc/test-plan.yaml' failed validation (document 1): invalid type 'bogus': must be one of 'functional', 'solution', 'performance', 'reliability', or 'security'
```

## 5. Generate output

```bash
# Gherkin feature files (default format)
gherkinator generate --format gh charmed-hpc/test-plan.yaml -o charmed-hpc
# -> charmed-hpc/gpu_job_submission.feature

# Markdown docs
gherkinator generate --format md charmed-hpc/test-plan.yaml -o charmed-hpc

# Cumulative risk filter: edge + beta plans only
gherkinator generate --format md charmed-hpc/test-plan.yaml -o charmed-hpc --risk beta

# Status filter: only implemented plans
gherkinator generate --format gh charmed-hpc/test-plan.yaml -o charmed-hpc --status implemented

# Tag filter: only plans tagged multi-node (untagged plans excluded)
gherkinator generate --format gh charmed-hpc/test-plan.yaml -o charmed-hpc --tag multi-node

# Union across several tags: plans tagged single-node OR minimal
gherkinator generate --format gh charmed-hpc/test-plan.yaml -o charmed-hpc --tag single-node --tag minimal

# Combine filters: edge-risk plans tagged multi-node
gherkinator generate --format gh charmed-hpc/test-plan.yaml -o charmed-hpc --risk edge --tag multi-node

# Mix files and directories; scan CWD when no args given
gherkinator generate --format md plans/ extras/another.yaml -o out
gherkinator generate --format gh -o out
```

Generated `.feature` files begin with a feature-level tag line that every
scenario inherits:

```gherkin
@functional @stable @planned @single-node
Feature: GPU job submission
```

## 6. Serve live docs

```bash
# Default project name (current working directory name)
gherkinator serve charmed-hpc/test-plan.yaml

# Override the project name shown in the docs
gherkinator serve charmed-hpc/test-plan.yaml --name "Charmed HPC"

# Serve only edge and beta risk plans
gherkinator serve charmed-hpc/test-plan.yaml --risk beta

# Serve only plans tagged for multi-node deployments
gherkinator serve charmed-hpc/test-plan.yaml --tag multi-node

# Serve from a directory of YAML files
gherkinator serve charmed-hpc/plans/ --name "Charmed HPC"
```

The command clones the Canonical Sphinx stack into `.gherkindocs/`,
generates Markdown, and runs `make run` in a TUI. Saving any input YAML
file triggers a rebuild. Stop the server with Ctrl+C. Requires `git`,
`python3`, `pip`, and `make`.

## 7. Delete plans

```bash
# Interactive confirmation
gherkinator delete -i test-plan.yaml "GPU job submission"

# Multiple plans, no confirmation
gherkinator delete -y -i test-plan.yaml "GPU job submission" "MPI job submission"
```

Feature-name matching is case-insensitive. If nothing matches, the file is
left untouched.

## 8. Clean generated artifacts

```bash
gherkinator clean                 # current directory
gherkinator clean -d charmed-hpc  # specific directory
```

Removes generated `.feature`/`.md` files and the `.gherkindocs/`
directory.

## 9. CI drift check with diff

```bash
# Fail if committed .feature files don't match the YAML source
gherkinator diff plans/ features/

# Only implemented plans, printing unified diffs for mismatches
gherkinator diff --status=implemented --show plans/ features/

# Restrict to plans at or below the beta risk level
gherkinator diff --risk=beta plans/ features/

# Restrict to plans tagged multi-node (repeat --tag for a union)
gherkinator diff --tag multi-node plans/ features/
```

Exit codes for CI scripting:

- `0` — in sync (pass the check)
- `1` — missing, differing, or orphan feature files (fail the check)
- `2` — input error such as invalid schema or missing features directory

Example CI step:

```yaml
- name: Ensure feature files match test plans
  run: gherkinator diff plans/ features/
```

A typical remediation loop when `diff` reports drift:

```bash
gherkinator diff --show plans/ features/   # inspect the unified diff
gherkinator generate --format gh plans/ -o features/
git add features/ && git commit -m "Regenerate feature files from test plans"
```

Note that editing any rendered classification field — `type`, `risk`,
`status`, or `tags` — changes the generated `.feature` content (they all
appear on the feature-level tag line). After changing any of them,
regenerate and re-commit the feature files or the CI drift check will
fail.
