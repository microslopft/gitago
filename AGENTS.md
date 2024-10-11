# AGENTS.md

Instructions for AI coding agents working in this repository.

## What this project is

`gitago` is a small Go CLI. It copies regular files from `-src` into `-dst` (honoring `.gitignore`), spreading them across `-commits` git commits dated in the past, then creates one extra commit **today** with an empty `README.md`.

Do not treat this as a general-purpose git client. The only git writes belong in `internal/gitrepo` and are used to build a synthetic history in `dst`.

## Layout

- `cmd/gitago` — CLI flags and `main`
- `internal/backfill` — scan, partition, timestamps, copy, YAML config, orchestration
- `internal/gitrepo` — `git init` / `add` / `commit` with authored dates
- `gitago.yaml` — sample messages and committers
- Agent docs: `AGENTS.md` (canonical), `CLAUDE.md`, `GEMINI.md`, `.github/copilot-instructions.md`
- Cursor: `.cursor/rules/`, `.cursor/skills/`
- Claude Code: `.claude/rules/`, `.claude/skills/`

The only third-party module is `gopkg.in/yaml.v3`. Do not add others unless unavoidable.

## Invariants

- Every regular file under `src` must appear in `dst` and in `git ls-files`, except content inside `.git` and paths excluded by `.gitignore` (root or nested). A parent directory ignore wins over a later negation of a child.
- `.gitignore` files themselves are copied unless they match an ignore rule.
- Historical commit count equals `-commits`. Total commits equal `-commits + 1`.
- Historical author/committer dates are strictly increasing and strictly before `Now`.
- The last commit date equals `Now` and includes empty `README.md` at the destination root.
- Commit messages are random alphanumeric strings (length 8–32) unless `-config` provides `messages`; then a random list entry is used.
- Author/committer is `gitago <gitago@local.lol>` unless `-config` provides `committers`; then a random name+email pair is used for both author and committer. `GITAGO_AUTHOR_NAME` / `GITAGO_AUTHOR_MAIL` override YAML `committers` when either is set.
- File mtimes are in the past and not after that file's commit time.
- `dst` must start empty; `src` and `dst` must not be the same or nested.
- `-seed` makes timestamps, grouping, messages, and identities deterministic.

## Commands

```bash
go test ./...
go run ./cmd/gitago -src <dir> -dst <dir> -commits <n>
go build -o bin/gitago ./cmd/gitago
```

Git must be on `PATH` for integration tests and for the CLI.

## Coding rules

- Standard library plus `gopkg.in/yaml.v3`.
- Inject time via `Config.Now` and git via `Config.Git`; do not call `time.Now()` inside helpers that tests need to freeze (except seed fallback when `Seed == 0`).
- Wrap errors with `%w` and the failing path or git args.
- Table-driven tests for pure helpers; real `git` in `t.TempDir()` for `Run`.
- Skip git integration tests when `exec.LookPath("git")` fails.
- Do not commit, push, or force-push unless the user asks.
- Do not rewrite history of **this** repository to demonstrate gitago; run the tool against throwaway `dst` directories.

## Changing behavior

If you change commit counting, date windows, README handling, file selection, gitignore, or YAML config, update:

1. Tests in `internal/backfill` and `internal/gitrepo`
2. `README.md` (user-facing, English)
3. This file, `.cursor/rules/`, `.cursor/skills/`, `.claude/rules/`, `.claude/skills/`

Prefer extending `Config` over adding global flags in multiple packages.
