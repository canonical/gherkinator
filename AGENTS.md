## Overview

This project follows the [standard Go project layout](https://github.com/golang-standards/project-layout).
`snap/snapcraft.yaml` is the snap package definition. `justfile` is the task runner.

## Code style

Follow the [Go Style Guide](https://google.github.io/styleguide/go/guide), plus:

### Imports

Three groups, alphabetized (`go fmt` handles ordering): standard library,
third-party, gherkinator.

### Avoid one-line assign/test

```go
err := doStuff()
if err != nil {
    return err
}
```

Not:

```go
if err := doStuff(); err != nil {
    return err
}
```

### Doc comments

Every exported (capitalised) name needs a doc comment immediately preceding the
declaration with no intervening blank lines.

## Build commands

```bash
just build      # builds bin/gherkinator
just install    # builds then installs to system GOPATH
just clean       # Remove build artifacts, .gherkindocs, and snap outputs.
```

### Building the snap package

`gherkinator` is distributed as a **classic** confined Snap on `core24` using
the `go` build plugin (`snap/snapcraft.yaml`).

```bash
just snap        # Build the snap with `snapcraft`.
```

- Do __not__ install the built snap on the host system.
- Do __not__ attempt to register or publish the snap to the Snap Store.
- If `snapcraft` encounters a Go build error, fix the Go source and tests first;
  the snapcraft YAML is unlikely to be the cause.

## Testing

```bash
just check            # Run all static checks (`lint` and `vet`).
just test unit        # Run unit tests with coverage profile (coverage.out).
just test             # Run all tests.
just coverage         # HTML coverage report (opens in browser).
just fmt              # Format all Go source code.
just lint             # `golangci-lint` static analysis.
just vet              # Vet all Go source code.
```

## Development workflow

1. Write Go code and matching `*_test.go` files alongside each source file.
2. Run unit tests: `just unit` — ensure 100% coverage.
3. Fix any test failures.
4. Format: `just fmt`.
5. Lint: `just lint` (uses `golangci-lint`). Fix all linter errors.
6. Vet: `just vet` (uses `go vet`). Fix all vetting errors.
7. Repeat for each new or modified file.

Pipe to the output of `just` to either `head` or `tail` to capture the `stdout` and/or `stderr`.

## Commit conventions

- Commits must be signed off (`Signed-off-by:` trailer) **by the human**. Agents
  must never add a `Signed-off-by:` trailer on the human's behalf.
- Agents must include an `Assisted-by:` trailer identifying the agent and model.
- Order trailers as: `Assisted-by:` first, then the human's `Signed-off-by:` last
  (added by the human).

Format:

    Assisted-by: AGENT_NAME:MODEL_VERSION:[MODEL_VARIANT]

- `AGENT_NAME`: The AI tool (for example, `opencode`).
- `MODEL_VERSION`: The specific model version used.
- `MODEL_VARIANT`: The variant of the model version used (for example, `low`, `medium`, or `high`). Optional

Other rules:

- Commit messages must be ASCII only.
- Keep PRs small and focused.
- Maintain a linear git history.

### Constraints

- Do **not** add new dependencies beyond what is already in `go.mod` without
  approval.
- Do **not** install anything with `apt` or `snap`.
- Do **not** run commands that require sudo.
- All errors must be handled explicitly in Go code.
