# Contributing to gherkinator

Do you want to contribute to gherkinator? You've come to
the right place then! __Here is how you can get involved.__

Please take a moment to review this document so that the contribution
process will be easy and effective for everyone. Following these guidelines
helps you communicate that you respect the maintainers and contributors
developing gherkinator. In return, they will reciprocate that respect
while addressing your issue or assessing your submitted changes and/or features.

### Table of Contents

* [Using the issue tracker](#using-the-issue-tracker)
* [Bug Reports](#bug-reports)
* [Enhancement Proposals](#enhancement-proposals)
* [Guidelines and Resources](#guidelines-and-resources)
* [Pull Requests](#pull-requests)

## Using the issue tracker

The issue tracker is the preferred way for tracking [bug reports](#bug-reports),
[enhancement proposals](#enhancement-proposals), and
[submitted pull requests](#pull-requests), but please follow these guidelines:

* Please __do not__ use the issue tracker for personal issues and/or support
  requests.

* Please __do not__ derail or troll issues. Keep the discussion on track and
  have respect for the other users and contributors.

* Please __do not__ post comments consisting solely of "+1", ":thumbsup:", or
  something similar. Use
  [GitHub's "reactions" feature](https://blog.github.com/2016-03-10-add-reactions-to-pull-requests-issues-and-comments/)
  instead.
  * The maintainers reserve the right to delete comments that violate this rule.

* Please __do not__ repost or reopen issues that have been closed. Please either
  submit a new issue or browse through previous issues.
  * The maintainers reserve the right to delete issues that violate this rule.

## Bug Reports

Guidelines for reporting bugs:

1. __Validate your issue__ &mdash; ensure that your issue is not being caused by
   a semantic or syntactic error in your own environment.

1. __Use the GitHub issue search__ &mdash; check if the issue has already been
   reported by someone else.

1. __Check if the issue has already been fixed__ &mdash; try to reproduce your
   issue using the latest revision of the repository.

1. __Isolate the problem__ &mdash; the more pinpointed the issue, the easier it
   is to fix.

A good bug report should not leave others needing to chase you for more
information. Please try to include answers to these questions:

* What is your current environment (OS, Go version, Snap version)?
* Which commands/flags/YAML input reproduce the issue?
* What was your expected outcome?
* What did you observe instead?

## Enhancement Proposals

The maintainers may already know what they want to add to gherkinator, but they
are always open to new ideas and potential improvements. More focused enhancement
discussions can start directly in an issue.

Please note that not all proposals may be incorporated into gherkinator. Spamming
the maintainers to incorporate something you want will not improve the likelihood
of it being implemented; it may result in a temporary ban.

## Guidelines and Resources

The following guidelines apply to all code contributions.

### Development environment

You will need:

| Tool | Purpose |
| --- | --- |
| Go 1.26+ | compiling and testing |
| [`just`](https://github.com/casey/just) | task runner |
| [`golangci-lint`](https://golangci-lint.run/) | linting |
| `snapcraft` | building the Snap package (optional) |

Clone the repository and verify everything works before making changes:

```bash
git clone https://github.com/canonical/gherkinator.git
cd gherkinator
just unit
just lint
```

### Go code

To have your Go code contributions accepted you must:

* Follow the [standard Go project layout](https://github.com/golang-standards/project-layout)
  and idiomatic Go style.

* Write unit tests using [`testify`](https://github.com/stretchr/testify) in a
  corresponding `<filename>_test.go` file for every new `<filename>.go` file.

* Achieve 100% unit test coverage. Run `just unit` to verify.

* Pass all linter checks. Run `just lint` and fix every reported issue before
  opening a pull request.

* Format the code with `just fmt` before committing.

* Handle all errors explicitly — do not ignore returned errors.

* Do not add new dependencies that are not already specified in
  [`.humans/plan.md`](.humans/plan.md) without first discussing the addition
  in an issue.

The recommended development loop for each new file is:

1. Write `<filename>.go` according to the implementation plan.
2. Write `<filename>_test.go` with comprehensive `testify` tests.
3. Run `just unit` — fix failures and ensure 100% coverage.
4. Run `just fmt` to format the code.
5. Run `just lint` — fix all reported issues.
6. Repeat until both commands report success.

### Snap package

The Snap configuration lives in `snap/snapcraft.yaml`. When updating it:

1. Keep it in sync with the implementation plan's Phase 12 description.
2. Build with `just snap` (or `snapcraft`) to verify it compiles correctly.
3. If you encounter Go compilation errors inside the Snapcraft build
   environment, fix them in the Go source first (following the Go code
   guidelines above) before retrying the snap build.

### Conventional Commits

* Follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
  for all commit messages.

* You can use Git's
  [interactive rebase](https://help.github.com/articles/about-git-rebase/)
  to tidy up commits before opening a pull request.

Common commit types used in this project:

| Type | When to use |
| --- | --- |
| `feat` | a new command, flag, or transpilation feature |
| `fix` | a bug fix |
| `test` | adding or updating tests |
| `docs` | changes to `README.md`, `CONTRIBUTING.md`, or inline comments |
| `chore` | dependency updates, `justfile` changes, CI config |
| `refactor` | code restructuring with no behaviour change |

### Useful resources

* [gherkinator `README.md`](README.md) — installation, schema reference, and
  command usage
* [`.humans/plan.md`](.humans/plan.md) — full implementation plan with
  annotated code for each phase
* [Go documentation](https://go.dev/doc/)
* [Bubbletea documentation](https://github.com/charmbracelet/bubbletea)
* [Cobra documentation](https://github.com/spf13/cobra)
* [Snapcraft documentation](https://snapcraft.io/docs)

## Pull Requests

Good pull requests — patches, improvements, new features — are a huge help.

__Ask first__ before embarking on any __significant__ pull request such as
implementing new features, refactoring methods, or incorporating new libraries;
otherwise, you risk spending a lot of time working on something that the
maintainers may not want to merge. For trivial changes or contributions that do
not require a large amount of time, you can go ahead and open a pull request.

Adhering to the following process is the best way to get your contribution
accepted:

1. [Fork](https://help.github.com/articles/fork-a-repo/) the project, clone
   your fork, and configure the remotes:

   ```bash
   # Clone your fork of the repo into the current directory
   git clone https://github.com/<your-username>/gherkinator.git

   # Navigate to the newly cloned directory
   cd gherkinator

   # Assign the original repo to a remote called "upstream"
   git remote add upstream https://github.com/canonical/gherkinator.git
   ```

2. If you cloned a while ago, pull the latest changes from upstream:

   ```bash
   git checkout main
   git pull upstream main
   ```

3. Create a new topic branch off `main`:

   ```bash
   git checkout -b <topic-branch-name>
   ```

4. Make your changes, then ensure all tests and linter checks pass:

   ```bash
   just unit
   just lint
   ```

5. Sign and commit your changes in logical chunks using
   [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

   To set up GPG or SSH key signing with git, see
   [GitHub's commit signature verification documentation](https://docs.github.com/en/authentication/managing-commit-signature-verification/about-commit-signature-verification).

6. Rebase the upstream development branch into your topic branch:

   ```bash
   git pull --rebase upstream main
   ```

7. Push your topic branch up to your fork:

   ```bash
   git push origin <topic-branch-name>
   ```

8. [Open a Pull Request](https://help.github.com/articles/about-pull-requests/)
   with a clear title and description against the `main` branch. The pull
   request should be focused — do not include commits unrelated to your
   contribution.

