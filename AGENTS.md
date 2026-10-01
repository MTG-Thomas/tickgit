# tickgit fork guidance

This Go fork turns latent-work comments into read-only reports and issue-candidate Markdown. It intentionally does not replace an issue tracker or create issues automatically. Read [README.md](README.md) and [docs/API.md](docs/API.md) before altering output or library behavior.

## Code map

`cmd/tickgit/commands/` contains CLI flags, TODO reports, statistics, and candidates. `pkg/comments/` handles language parsing and ignore behavior; `pkg/blame/` handles history; `pkg/todos/`, `pkg/stats/`, `pkg/baseline/`, and `pkg/issuecandidates/` own reusable reporting and comparisons. Fixtures under `testdata/repos/` and package testdata are part of the regression contract, not disposable working repositories.

## Verification

Follow the Go version declared in `go.mod` and workflows. CI runs `go vet -v ./...` and `go test -v ./...`; release also runs `go test ./...`. Preserve tests for phrase overrides, default ignores, colors/NO_COLOR, blame warnings, baseline comparisons, and stable candidate keys when changing those seams. Source inspection of a command is not an executed check.

## Guard and release boundaries

`action.yml` implements a read-only latent-work guard; `.github/tickgit-baseline.csv` records accepted existing findings. Keep read-only permissions and release-pinned Action consumption. Changes to matching or normalization must explain their baseline impact; do not refresh a baseline simply to hide new findings. Candidate generation remains a human-review step before issue creation.

Inspect `.github/workflows/release.yml` and `.goreleaser.yml` before release changes. Tagging, publishing assets, modifying another repo's baseline, and posting GitHub issues are separate authorized actions. Preserve upstream attribution while keeping MTG fork behavior and URLs accurate.
