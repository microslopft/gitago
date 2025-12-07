package backfill

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type yamlFile struct {
	Messages   []string        `yaml:"messages"`
	Committers []yamlCommitter `yaml:"committers"`
}

type yamlCommitter struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

// LoadYAML reads optional commit messages and committer identities.
// Empty lists mean the caller should keep the built-in random behavior.
func LoadYAML(path string) (messages []string, committers []Identity, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read config: %w", err)
	}
	var raw yamlFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", path, err)
	}
	for i, msg := range raw.Messages {
		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue
		}
		if strings.Contains(msg, "\n") {
			return nil, nil, fmt.Errorf("%s: messages[%d] must be a single line", path, i)
		}
		messages = append(messages, msg)
	}

	aName := os.Getenv("GITAGO_AUTHOR_NAME")
	aMail := os.Getenv("GITAGO_AUTHOR_MAIL")

	if aName == "" && aMail == "" {
		for i, c := range raw.Committers {
			name := strings.TrimSpace(c.Name)
			email := strings.TrimSpace(c.Email)
			if name == "" && email == "" {
				continue
			}
			if name == "" || email == "" {
				return nil, nil, fmt.Errorf("%s: committers[%d] needs both name and email", path, i)
			}
			if strings.ContainsAny(email, " \t") {
				return nil, nil, fmt.Errorf("%s: committers[%d] email is invalid", path, i)
			}
			committers = append(committers, Identity{Name: name, Email: email})
		}
	} else {
		committers = append(committers, Identity{Name: aName, Email: aMail})
	}

	return messages, committers, nil
}
