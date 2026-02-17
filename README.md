# GitAgo

Copies files from a source tree into a new git repository, spreading them across backdated commits, and finishes today with an empty `README.md`.

## Requirements

- Go 1.22+
- `git` on `PATH`

## Build and run

```bash
go run ./cmd/gitago -src ./testdata/src -dst ./out -commits 12
```

```bash
go build -o bin/gitago ./cmd/gitago
./bin/gitago -src ./src -dst ./dst -commits 20 -days 180 -seed 42 -config gitago.yaml
```

### Flags

| Flag | Description |
|------|-------------|
| `-src` | Source directory (required) |
| `-dst` | Empty destination directory; created if needed (required) |
| `-commits` | Number of **past** commits that receive source files (required, ≥ 1) |
| `-days` | How far back the first commit may be, in days (default 365) |
| `-seed` | RNG seed; `0` uses the current time |
| `-config` | YAML with `messages` and `committers`; if set, values are picked at random instead of a random string / `gitago@local.lol` |

After `N` historical commits, one more is always added: today, with an empty `README.md` at the `dst` root. The repo ends up with `N + 1` commits.

## Behavior

1. Regular files are collected recursively from `-src`. The `.git` directory is skipped. Paths matched by `.gitignore` (root or nested) are not copied to `dst`.
2. A local repository is created in `-dst` on branch `master`.
3. Files are assigned to `N` commits in random order. If there are more commits than files, some commits are empty (`--allow-empty`).
4. Copied files get random modification times in the past (and creation times on macOS if `SetFile` is available), not later than their commit date.
5. Commit messages are random alphanumeric strings. If `-config` provides `messages`, a random entry is used. The default author is `gitago <gitago@local.lol>`; if `committers` is non-empty, a random name+email pair is used. `GITAGO_AUTHOR_NAME` / `GITAGO_AUTHOR_MAIL` override YAML committers when either is set.
6. Commit dates are strictly increasing and fall in `[now − days, now)`.
7. The last commit is dated “now” and contains an empty `README.md`.

Sample file: [`gitago.yaml`](gitago.yaml) — common commit phrasings and a set of names.

## Tests

```bash
go test ./...
```

## For agents

Canonical project guide: [`AGENTS.md`](AGENTS.md).
