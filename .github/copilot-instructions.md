Follow AGENTS.md at the repository root.

gitago is a Go CLI that copies files from -src into a new git repo at -dst across N backdated commits (skipping .gitignore matches), then commits an empty README.md dated today.

Use go test ./... before finishing changes. The only extra module is gopkg.in/yaml.v3. Put git command execution only in internal/gitrepo.
