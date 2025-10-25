package gitrepo

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	authorName  = "gitago"
	authorEmail = "gitago@local.lol"
)

// Client runs git commands in a working tree.
type Client struct {
	// Bin is the git executable. Empty means "git" from PATH.
	Bin string
}

func (c Client) bin() string {
	if c.Bin != "" {
		return c.Bin
	}
	return "git"
}

func (c Client) run(dir string, env []string, args ...string) error {
	cmd := exec.Command(c.bin(), args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %v: %w\n%s", args, err, out)
	}
	return nil
}

// Init creates a new repository on branch master and sets a local identity
// so commits work without a global git config.
func (c Client) Init(dir string) error {
	if err := c.run(dir, nil, "init", "-b", "master"); err != nil {
		return err
	}
	if err := c.run(dir, nil, "config", "user.name", authorName); err != nil {
		return err
	}
	if err := c.run(dir, nil, "config", "user.email", authorEmail); err != nil {
		return err
	}
	return c.run(dir, nil, "config", "commit.gpgsign", "false")
}

// Add stages the given paths (relative to dir).
func (c Client) Add(dir string, paths ...string) error {
	if len(paths) == 0 {
		return nil
	}
	args := append([]string{"add", "--"}, paths...)
	return c.run(dir, nil, args...)
}

// Commit creates a commit at the given author/committer time.
// Empty name or email fall back to gitago/gitago@local.lol.
func (c Client) Commit(dir, message string, when time.Time, allowEmpty bool, name, email string) error {
	if name == "" {
		name = authorName
	}
	if email == "" {
		email = authorEmail
	}
	stamp := when.Format(time.RFC3339)
	env := []string{
		"GIT_AUTHOR_DATE=" + stamp,
		"GIT_COMMITTER_DATE=" + stamp,
		"GIT_AUTHOR_NAME=" + name,
		"GIT_AUTHOR_EMAIL=" + email,
		"GIT_COMMITTER_NAME=" + name,
		"GIT_COMMITTER_EMAIL=" + email,
	}
	args := []string{
		"-c", "user.name=" + name,
		"-c", "user.email=" + email,
		"-c", "commit.gpgsign=false",
		"commit", "-m", message,
	}
	if allowEmpty {
		args = append(args, "--allow-empty")
	}
	return c.run(dir, env, args...)
}
