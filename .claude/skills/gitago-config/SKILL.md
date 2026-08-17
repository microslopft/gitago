---
name: gitago-config
description: >-
  Edit or explain gitago YAML config (messages, committers) and author
  environment overrides. Use when the user mentions gitago.yaml, -config,
  commit messages, committer names/emails, GITAGO_AUTHOR_NAME, or
  GITAGO_AUTHOR_MAIL.
---

# gitago config

Pass `-config path.yaml`. Omitted file → random alphanumeric messages (8–32) and `gitago <gitago@local.lol>`.

## Schema

```yaml
messages:
  - 'fix: tweaks'
committers:
  - name: John Smith
    email: john.smith@mail.lol
```

- `messages`: single-line strings; empty entries skipped. Non-empty list → pick one at random per commit.
- `committers`: each entry needs both `name` and `email` (no spaces in email). Non-empty list → pick one at random; used as author **and** committer.
- Either list may be omitted; the other still applies.
- Shipped sample: [`gitago.yaml`](../../../gitago.yaml) (Conventional Commits + English text).

## Env override

If **both** `GITAGO_AUTHOR_NAME` and `GITAGO_AUTHOR_MAIL` are unset, YAML `committers` are used.

If **either** is set, YAML `committers` are ignored and that single identity is used (values taken as-is, including empty).

## Loader

`internal/backfill.LoadYAML`. Reject multiline messages and incomplete committer rows.

After schema changes, update `yaml_test.go` / `run_extra_test.go` and the matching skills/rules in `.claude` and `.cursor`.
