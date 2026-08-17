---
description: gitago project invariants and layout
---

# gitago

Canonical guide: `AGENTS.md`. Skills: `.claude/skills/` and `.cursor/skills/`.

- CLI `cmd/gitago` → logic `internal/backfill` → git I/O `internal/gitrepo` only.
- Module `github.com/microslopft/gitago`. Extra dep: `gopkg.in/yaml.v3` only.
- `-commits` = past commits; plus one **today** with empty `README.md`. `dst` branch is `master`.
- Copy all regular `src` files except `.git` and `.gitignore` matches. Parent-dir ignore wins over `!child`.
- No `-config`: random messages (8–32) and `gitago <gitago@local.lol>`.
- `-config` YAML: `messages` and/or `committers`. `GITAGO_AUTHOR_NAME` / `GITAGO_AUTHOR_MAIL` override YAML committers if either is set.
- Inject `Config.Now` and `Config.Git`. `-seed` makes dates, grouping, messages, identities deterministic.
- After behavior changes: tests, `README.md`, `AGENTS.md`, `.claude/rules/`, `.claude/skills/`, `.cursor/rules/`, `.cursor/skills/`.
- `go test ./...`. Skip git tests if `git` is missing. Do not demo on this repo's history.
