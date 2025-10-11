package backfill

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/microslopft/gitago/internal/gitrepo"
)

func TestRun_BackfillsHistory(t *testing.T) {
	requireGit(t)

	src := t.TempDir()
	dst := t.TempDir()
	writeFile(t, filepath.Join(src, "a.txt"), "alpha")
	writeFile(t, filepath.Join(src, "nested", "b.txt"), "beta")
	writeFile(t, filepath.Join(src, "c.bin"), "\x00\x01")

	now := time.Date(2026, 8, 16, 15, 21, 0, 0, time.UTC)
	if err := Run(Config{
		Src:     src,
		Dst:     dst,
		Commits: 4,
		Days:    40,
		Seed:    42,
		Now:     now,
		Git:     gitrepo.Client{},
	}); err != nil {
		t.Fatal(err)
	}

	assertFile(t, filepath.Join(dst, "a.txt"), "alpha")
	assertFile(t, filepath.Join(dst, "nested", "b.txt"), "beta")
	assertFile(t, filepath.Join(dst, "c.bin"), "\x00\x01")

	readme, err := os.ReadFile(filepath.Join(dst, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(readme) != 0 {
		t.Fatalf("README.md must be empty, got %q", readme)
	}

	count := strings.TrimSpace(gitOutput(t, dst, "rev-list", "--count", "HEAD"))
	if count != "5" {
		t.Fatalf("commit count=%s, want 5", count)
	}

	dates := strings.Split(strings.TrimSpace(gitOutput(t, dst, "log", "--reverse", "--format=%aI")), "\n")
	if len(dates) != 5 {
		t.Fatalf("dates=%v", dates)
	}
	parsed := make([]time.Time, 0, len(dates))
	for i, raw := range dates {
		tm, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			t.Fatal(err)
		}
		parsed = append(parsed, tm)
		if i < len(dates)-1 && !tm.Before(now) {
			t.Fatalf("historical commit %d not in the past: %v", i, tm)
		}
	}
	last := parsed[len(parsed)-1]
	if !last.Equal(now) {
		t.Fatalf("last commit %v, want %v", last, now)
	}
	for i := 1; i < len(parsed); i++ {
		if !parsed[i].After(parsed[i-1]) {
			t.Fatalf("commit dates not increasing: %v", parsed)
		}
	}

	files := gitOutput(t, dst, "ls-files")
	for _, name := range []string{"a.txt", filepath.Join("nested", "b.txt"), "c.bin", "README.md"} {
		if !strings.Contains(files, name) {
			t.Fatalf("git is missing %s:\n%s", name, files)
		}
	}

	headFiles := gitOutput(t, dst, "show", "--name-only", "--pretty=format:", "HEAD")
	if !strings.Contains(headFiles, "README.md") {
		t.Fatalf("HEAD should include README.md, got %q", headFiles)
	}

	msgs := strings.Split(strings.TrimSpace(gitOutput(t, dst, "log", "--format=%s")), "\n")
	if len(msgs) != 5 {
		t.Fatalf("messages=%v", msgs)
	}
	uniq := map[string]struct{}{}
	for _, m := range msgs {
		if len(m) < 8 || len(m) > 32 {
			t.Fatalf("random message length %q", m)
		}
		for _, r := range m {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
				t.Fatalf("random message %q", m)
			}
		}
		uniq[m] = struct{}{}
	}
	if len(uniq) < 2 {
		t.Fatalf("expected random messages, got %v", msgs)
	}

	authors := strings.Split(strings.TrimSpace(gitOutput(t, dst, "log", "--format=%an <%ae>%n%cn <%ce>")), "\n")
	for _, a := range authors {
		if a != "gitago <gitago@local.lol>" {
			t.Fatalf("default identity %q", a)
		}
	}

	start := now.AddDate(0, 0, -40)
	for _, rel := range []string{"a.txt", filepath.Join("nested", "b.txt"), "c.bin"} {
		info, err := os.Stat(filepath.Join(dst, rel))
		if err != nil {
			t.Fatal(err)
		}
		mt := info.ModTime()
		if mt.After(now) {
			t.Fatalf("%s mtime %v is not in the past", rel, mt)
		}
		if mt.Before(start.Add(-time.Second)) {
			t.Fatalf("%s mtime %v is before window start %v", rel, mt, start)
		}
	}
}

func TestRun_DeterministicSeed(t *testing.T) {
	requireGit(t)
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)

	logOf := func() string {
		src := t.TempDir()
		dst := t.TempDir()
		writeFile(t, filepath.Join(src, "one.txt"), "1")
		writeFile(t, filepath.Join(src, "two.txt"), "2")
		if err := Run(Config{Src: src, Dst: dst, Commits: 3, Days: 20, Seed: 7, Now: now}); err != nil {
			t.Fatal(err)
		}
		return gitOutput(t, dst, "log", "--reverse", "--format=%aI %s %an <%ae> %cn <%ce>")
	}

	if a, b := logOf(), logOf(); a != b {
		t.Fatalf("seed was not deterministic\nA:\n%s\nB:\n%s", a, b)
	}
}

func TestRun_RejectsNonEmptyDst(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	writeFile(t, filepath.Join(src, "a.txt"), "a")
	writeFile(t, filepath.Join(dst, "keep.txt"), "nope")
	err := Run(Config{Src: src, Dst: dst, Commits: 1, Days: 10, Seed: 1, Now: time.Now()})
	if err == nil || !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("got %v", err)
	}
}

func TestRun_RejectsEmptySrc(t *testing.T) {
	err := Run(Config{
		Src:     t.TempDir(),
		Dst:     t.TempDir(),
		Commits: 1,
		Days:    10,
		Seed:    1,
		Now:     time.Now(),
	})
	if err == nil || !strings.Contains(err.Error(), "no files") {
		t.Fatalf("got %v", err)
	}
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("%s: got %q want %q", path, got, want)
	}
}
