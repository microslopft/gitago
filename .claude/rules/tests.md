---
description: How to write gitago tests
paths:
  - "**/*_test.go"
---

# Tests

- Table-driven helpers; real `git` in `t.TempDir()` for `Run`. Skip if `exec.LookPath("git")` fails.
- Freeze `Config.Now`, set `Config.Seed`. Inject `Messages` / `Committers` or `LoadYAML("../../gitago.yaml")`.
- Assert `rev-list --count HEAD` == `commits+1`, branch `master`, HEAD has empty `README.md`.
- Default identity in code is `gitago <gitago@local.lol>` — keep log assertions in sync.
- Cover messages-only, committers-only, gitignore skip, missing `dst`, ignored-all `src`.
- `go test ./...` and `go test -race -count=1 ./...` before finishing.
