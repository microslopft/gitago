---
name: gitago-tests
description: >-
  Write and update gitago Go tests for backfill, gitignore, YAML config, and
  gitrepo. Use when adding tests, fixing failing tests, or changing behavior
  that must stay covered.
---

# gitago tests

```bash
go test ./...
go test -race -count=1 ./...
```

Skip git integration tests when `exec.LookPath("git")` fails.

## Where tests live

| Area | Files |
|------|--------|
| CLI flags and `run()` | `cmd/gitago/main_test.go` |
| `Config.normalize` | `internal/backfill/config_test.go` |
| `Run` history | `internal/backfill/run_test.go`, `run_extra_test.go` |
| `.gitignore` | `internal/backfill/gitignore_test.go` |
| YAML | `internal/backfill/yaml_test.go` |
| dates / pickers | `internal/backfill/schedule_test.go` |
| copy / collect | `internal/backfill/files_test.go` |
| git subprocess | `internal/gitrepo/git_test.go` |

## Rules

- Table-driven tests for pure helpers; real `git` + `t.TempDir()` for `Run`.
- Freeze `Config.Now` and set `Config.Seed`.
- Inject `Config.Messages` / `Config.Committers`, or `LoadYAML("../../gitago.yaml")`.
- Default identity in **code** is `gitago <gitago@local.lol>`. Keep assertions in sync.
- Destination branch is `master`.
- Assert `rev-list --count HEAD` == `commits+1` and `ls-files` has every non-ignored source file plus `README.md`.
- Cover: messages-only, committers-only, gitignore skip, missing `dst`, empty `README.md` on HEAD.

Do not commit unless the user asks.
