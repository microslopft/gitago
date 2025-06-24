package backfill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Git writes a local repository in the destination tree.
type Git interface {
	Init(dir string) error
	Add(dir string, paths ...string) error
	Commit(dir, message string, when time.Time, allowEmpty bool, name, email string) error
}

// Identity is a git author/committer.
type Identity struct {
	Name  string
	Email string
}

// Config is the backfill run parameters.
type Config struct {
	Src        string
	Dst        string
	Commits    int
	Days       int
	Seed       int64
	Now        time.Time
	Git        Git
	Messages   []string
	Committers []Identity
}

func (c *Config) normalize() error {
	if c.Src == "" {
		return fmt.Errorf("src is required")
	}
	if c.Dst == "" {
		return fmt.Errorf("dst is required")
	}
	if c.Commits < 1 {
		return fmt.Errorf("commits must be >= 1")
	}
	if c.Days < 1 {
		return fmt.Errorf("days must be >= 1")
	}

	src, err := filepath.Abs(c.Src)
	if err != nil {
		return fmt.Errorf("src: %w", err)
	}
	dst, err := filepath.Abs(c.Dst)
	if err != nil {
		return fmt.Errorf("dst: %w", err)
	}
	c.Src, c.Dst = src, dst

	if sameOrNested(src, dst) {
		return fmt.Errorf("src and dst must be distinct directories")
	}

	st, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("src: %w", err)
	}
	if !st.IsDir() {
		return fmt.Errorf("src %q is not a directory", src)
	}

	if c.Now.IsZero() {
		c.Now = time.Now()
	}
	return nil
}

func sameOrNested(a, b string) bool {
	return within(a, b) || within(b, a)
}

func within(parent, path string) bool {
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
