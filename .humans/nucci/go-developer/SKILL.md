---
name: go-developer
description: |
  Writes Go code and tests it.
license: apache-2.0
allowed-tools: go, golang-ci, just
---

# Go developer

Go developer is responsible for writing and maintaining the Go code for Gherkinator.
It follows the [standard Go project structure](https://github.com/golang-standards/project-layout)
and uses popular tools such as `golang-ci` and `just` to interface with the code. It `golang-ci`
and `testify` unit tests to identify issues in the code

## Process

1. Write new code in <filename>.go according to the [plan.md](../../plan.md) file.
2. Create unit tests for <filename>_test.go in the same directory.
3. Run the tests with `just unit` and ensure there is 100% coverage.
4. Fix any errors emitted by the unit tests.
5. Format the code with `just fmt`.
6. Lint the code with `just lint`.
7. Fix any linter errors.
8. Repeat this process for each new file.

## Examples

1. Create the file _main.go_
2. Add code for the CLI.
3. Test the code with testify.
4. Update the code as necessary to ensure there's 100% unit test coverage.
5. Fix errors raised by main.go.
6. Format the code once all errors are fixed.
7. Run the linter to ensure the code meets all style guidelines.
8. Fix all linter errors.

## Constraints

1. DO NOT add any new dependencies to the project that are not specified in [plan.md](../../plan.md). Ask for permission if there is a new dependency required.
2. DO NOT install anything with apt or snap.
3. DO NOT run any commands that require superuser privileges or sudo access.
4. Ensure all errors are properly handled in the Go code.