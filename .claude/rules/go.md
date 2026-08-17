---
description: Go conventions for gitago
paths:
  - "**/*.go"
---

# Go

Wrap git errors with `%w` and context. No panics in library code; validate in `Config.normalize`.

- Stdlib plus `gopkg.in/yaml.v3`. YAML parsing stays in `internal/backfill/yaml.go`.
- `math/rand/v2` + `rand.NewPCG` from `Config.Seed`.
- `filepath` for disk paths; pass repo-relative paths to `git add`.
- Default identity constants live in `internal/gitrepo` (`gitago`, `gitago@local.lol`). Init `-b master`.
