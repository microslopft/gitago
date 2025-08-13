# CLAUDE.md

Instructions for Claude Code in this repository.

Long-form guide: [`AGENTS.md`](AGENTS.md).  
Rules: [`.claude/rules/`](.claude/rules/). Skills: [`.claude/skills/`](.claude/skills/).

## Skills

| Skill | Use when |
|-------|----------|
| [gitago-backfill](.claude/skills/gitago-backfill/SKILL.md) | Running or changing the backfill CLI |
| [gitago-config](.claude/skills/gitago-config/SKILL.md) | YAML `messages`/`committers` or author env vars |
| [gitago-tests](.claude/skills/gitago-tests/SKILL.md) | Adding or updating tests |

## Facts

- Module: `github.com/microslopft/gitago`
- Entry: `cmd/gitago` → `internal/backfill` → `internal/gitrepo`
- Extra module: `gopkg.in/yaml.v3` only
- Destination repo branch: `master`
- Default identity: `gitago <gitago@local.lol>`
- `-commits` = past commits; plus one commit **today** with empty `README.md`
- `.gitignore` in `src` is honored; `.git` is skipped
- `GITAGO_AUTHOR_NAME` / `GITAGO_AUTHOR_MAIL` override YAML `committers` if either is set

## Commands

```bash
go test ./...
go run ./cmd/gitago -src <dir> -dst <dir> -commits <n> [-days 365] [-seed 0] [-config gitago.yaml]
go build -o bin/gitago ./cmd/gitago
```

Do not rewrite history of this repo to demo the tool. Use a throwaway `-dst`.
