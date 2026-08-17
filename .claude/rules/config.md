---
description: gitago.yaml and author env overrides
paths:
  - "gitago.yaml"
  - "**/yaml.go"
  - "**/yaml_test.go"
---

# Config

`-config` loads YAML via `backfill.LoadYAML`.

```yaml
messages:
  - 'fix: tweaks'
committers:
  - name: John Smith
    email: john.smith@mail.lol
```

- Messages are single-line; blanks skipped. Committers need both name and email (no spaces in email).
- Empty/omitted lists keep random messages and `gitago@local.lol`.
- If `GITAGO_AUTHOR_NAME` or `GITAGO_AUTHOR_MAIL` is set, YAML `committers` are ignored; that one identity is used as-is.
- Shipped sample: `gitago.yaml` (Conventional Commits + English).
