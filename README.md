[![tests](https://github.com/MTG-Thomas/tickgit/actions/workflows/test.yml/badge.svg)](https://github.com/MTG-Thomas/tickgit/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/MTG-Thomas/tickgit)](https://goreportcard.com/report/github.com/MTG-Thomas/tickgit)
[![release](https://img.shields.io/github/v/release/MTG-Thomas/tickgit)](https://github.com/MTG-Thomas/tickgit/releases)
[![license](https://img.shields.io/github/license/MTG-Thomas/tickgit)](LICENSE)

## tickgit 🎟️

This fork keeps the original `tickgit` idea, but points it at a more opinionated
goal: make latent work in code visible enough to promote into proper GitHub
issues, and prevent new latent-work comments from quietly accumulating in local
source.

Use the `tickgit` command to view pending tasks, progress reports, completion
summaries, historical data from `git` history, and issue-candidate markdown.
Use the included GitHub Action as a read-only latent-work guard on pull requests
and schedules.

This fork is not meant to replace GitHub Issues, JIRA, Trello, or other project
management tools. It is meant to catch the work that starts as comments,
tickets, or checklists in a codebase and give maintainers a safe path to either
track it properly or stop adding more of it.

### Fork goals

- Detect latent-work comments locally and in CI without mutating repositories.
- Compare current findings to a committed baseline so existing latent work does
  not block adoption.
- Fail pull requests or scheduled checks only when new findings appear.
- Generate issue-candidate markdown for human review before anything becomes a
  GitHub issue.
- Keep supply-chain risk low by using read-only permissions, release-pinned
  Action refs, and explicit baselines.

### TODOs

`tickgit` will scan a codebase and identify any TODO items in the comments. It will output a report like so:

```
# tickgit ~/Desktop/facebook/react
...
TODO:
  => packages/scheduler/src/__tests__/SchedulerBrowser-test.js:85:9
  => added 1 month ago by Andrew Clark <git@andrewclark.io> in a2e05b6c148b25590884e8911d4d4acfcb76a487

TODO: Scheduler no longer requires these methods to be polyfilled. But
  => packages/scheduler/src/__tests__/SchedulerBrowser-test.js:77:7
  => added 1 month ago by Andrew Clark <git@andrewclark.io> in a2e05b6c148b25590884e8911d4d4acfcb76a487

TODO: Scheduler no longer requires these methods to be polyfilled. But
  => packages/scheduler/src/forks/SchedulerHostConfig.default.js:77:7
  => added 1 month ago by Andrew Clark <git@andrewclark.io> in a2e05b6c148b25590884e8911d4d4acfcb76a487

TODO: useTransition hook instead.
  => fixtures/concurrent/time-slicing/src/index.js:110:11
  => added 3 weeks ago by Sebastian Markbåge <sebastian@calyptus.eu> in 3ad076472ce9108b9b8a6a6fe039244b74a34392

128 TODOs Found 📝
```

Check out [an example](https://www.tickgit.com/browse?repo=github.com/kubernetes/kubernetes) of the TODOs tickgit will surface for the Kubernetes codebase.

#### Coming Soon

- [x] Blame - get a better sense of how old TODOs are, when they were introduced and by whom
- [x] Context - use `--context-lines <n>` for visibility into the lines of code _around_ a TODO
- [x] More `TODO` type phrases to match, such as `FIXME`, `XXX`, `HACK`, or customized alternatives.
- [x] More configurability (e.g. custom ignore paths and color output)
- [x] Markdown parsing
- [x] More thorough historical stats

### Installation

#### GitHub Releases

Download prebuilt binaries from the [latest release](https://github.com/MTG-Thomas/tickgit/releases/latest).

### Usage

The most up to date usage will be the output of `tickgit --help`.

#### Historical stats

Use `tickgit stats` to summarize current findings by phrase, age bucket, and
directory, plus the oldest findings according to Git blame metadata.

```sh
tickgit stats
tickgit stats --json
```

#### Match phrases

By default, tickgit matches `TODO`, `FIXME`, `OPTIMIZE`, `HACK`, `XXX`, `WTF`,
and `LEGACY`, plus the `@lowercase` form for each phrase. Use
`--match-phrase` to override that list.

```sh
tickgit --match-phrase TODO --match-phrase FIXME
tickgit --match-phrase TODO,FIXME --csv-output
tickgit stats --match-phrase TODO --json
```

#### Ignore paths and color

Tickgit skips common repository metadata, dependency, and build paths by
default, including `.git`, `node_modules`, `vendor`, `dist`, `build`, `target`,
`bin`, `obj`, `.terraform`, virtual environment folders, and coverage output.
Use `--ignore-path` to add repository-specific patterns.

```sh
tickgit --ignore-path fixtures --ignore-path generated
tickgit stats --ignore-path docs/generated
```

Human-readable output is colorized by default unless `NO_COLOR` is set. Use
`--color auto`, `--color always`, or `--color never` to choose explicitly.

```sh
NO_COLOR=1 tickgit
tickgit --color never
```

#### Issue candidates

Use `tickgit candidates` to turn tickgit CSV output into issue-candidate
markdown with stable duplicate keys. This command is intended for scheduled
curation workflows; it does not create issues or mutate repositories.

```sh
tickgit --csv-output > .github/tickgit-current.csv
tickgit candidates --repo MTG-Thomas/tickgit --csv-file .github/tickgit-current.csv > .github/tickgit-candidates.md
```

### GitHub Action

This fork includes a read-only GitHub Action that compares current tickgit CSV
output to a committed baseline. It is intended for scheduled and pull request
checks that fail only when a repository introduces new latent-work comments.
The Action does not create issues, comments, commits, or other mutations.

Create a baseline:

```sh
tickgit --csv-output > .github/tickgit-baseline.csv
```

Then add a workflow:

```yaml
name: tickgit TODO guard

on:
  pull_request:
  schedule:
    - cron: "17 13 * * 1-5"
  workflow_dispatch:

permissions:
  contents: read

jobs:
  todo-guard:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: MTG-Thomas/tickgit@v0.0.17
        with:
          baseline-file: .github/tickgit-baseline.csv
          fail-on-new: "true"
          match-phrases: TODO,FIXME,HACK
          ignore-paths: fixtures,docs/generated
```

Prefer pinning to a release tag or reviewed commit SHA rather than `main` when
rolling the Action out across repositories. The Action builds tickgit from the
pinned Action ref and exits with status 2 when new findings appear relative to
the baseline.

### API

To find information about using the tickgit API, see [this file](https://github.com/MTG-Thomas/tickgit/blob/main/docs/API.md).
