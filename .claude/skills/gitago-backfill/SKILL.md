---
name: gitago-backfill
description: >-
  Run or extend the gitago CLI that copies src files into a new dst git repo
  with backdated commits and a final empty README.md today. Use when the user
  mentions gitago, synthetic git history, backdated commits, src/dst copy,
  gitignore, or CLI flags.
---

# gitago backfill

## Run

```bash
go run ./cmd/gitago -src <dir> -dst <dir> -commits <n> [-days 365] [-seed 0] [-config gitago.yaml]
```

`dst` must be empty or absent. Result: `n` past commits + 1 commit now with empty `README.md`. Branch in `dst` is `master`.

## Invariants

- Copy every regular file from `src` except `.git` and `.gitignore` matches.
- Nested `.gitignore` applies; an ignored parent directory wins over `!child`.
- `.gitignore` itself is copied unless ignored.
- If `commits` > file count, extra historical commits are empty (`--allow-empty`).
- Historical dates strictly increase and are before `Now`; last commit date is `Now`.
- File mtimes are in the past and not after that file's commit time.
- `src` and `dst` must not be the same or nested.
- `-seed` makes grouping, dates, messages, and identities deterministic.

## Change safely

1. Keep git subprocesses in `internal/gitrepo` only.
2. Freeze time with `Config.Now`; inject `Config.Git` in tests.
3. Do not add dependencies besides `gopkg.in/yaml.v3`.
4. After behavior changes: tests, `README.md`, `AGENTS.md`, `.claude/rules/`, `.claude/skills/`, `.cursor/rules/`, `.cursor/skills/`.

Details: [AGENTS.md](../../../AGENTS.md). Config: [gitago-config](../gitago-config/SKILL.md). Tests: [gitago-tests](../gitago-tests/SKILL.md).
