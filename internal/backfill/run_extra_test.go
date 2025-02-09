package backfill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRun_MoreCommitsThanFiles(t *testing.T) {
	requireGit(t)
	src := t.TempDir()
	dst := t.TempDir()
	writeFile(t, filepath.Join(src, "only.txt"), "one")
	now := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	if err := Run(Config{Src: src, Dst: dst, Commits: 5, Days: 15, Seed: 3, Now: now}); err != nil {
		t.Fatal(err)
	}
	count := strings.TrimSpace(gitOutput(t, dst, "rev-list", "--count", "HEAD"))
	if count != "6" {
		t.Fatalf("count=%s", count)
	}
	body, err := os.ReadFile(filepath.Join(dst, "only.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "one" {
		t.Fatalf("body=%q", body)
	}
}

func TestRun_RespectsGitignore(t *testing.T) {
	requireGit(t)
	src := t.TempDir()
	dst := t.TempDir()
	writeFile(t, filepath.Join(src, ".gitignore"), "*.log\nbuild/\n")
	writeFile(t, filepath.Join(src, "keep.txt"), "yes")
	writeFile(t, filepath.Join(src, "noise.log"), "no")
	writeFile(t, filepath.Join(src, "build", "out.go"), "no")

	now := time.Date(2026, 8, 16, 18, 0, 0, 0, time.UTC)
	if err := Run(Config{Src: src, Dst: dst, Commits: 2, Days: 10, Seed: 9, Now: now}); err != nil {
		t.Fatal(err)
	}

	assertFile(t, filepath.Join(dst, "keep.txt"), "yes")
	assertFile(t, filepath.Join(dst, ".gitignore"), "*.log\nbuild/\n")
	if _, err := os.Stat(filepath.Join(dst, "noise.log")); !os.IsNotExist(err) {
		t.Fatalf("ignored log was copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "build", "out.go")); !os.IsNotExist(err) {
		t.Fatalf("ignored build file was copied: %v", err)
	}

	listed := gitOutput(t, dst, "ls-files")
	for _, name := range []string{"noise.log", filepath.Join("build", "out.go")} {
		if strings.Contains(listed, name) {
			t.Fatalf("git still tracks ignored %s:\n%s", name, listed)
		}
	}
	if !strings.Contains(listed, "keep.txt") {
		t.Fatalf("keep.txt missing:\n%s", listed)
	}
}

func TestRun_CreatesMissingDst(t *testing.T) {
	requireGit(t)
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "a.txt"), "a")
	dst := filepath.Join(t.TempDir(), "nested", "repo")
	now := time.Date(2026, 8, 16, 11, 0, 0, 0, time.UTC)
	if err := Run(Config{Src: src, Dst: dst, Commits: 1, Days: 8, Seed: 2, Now: now}); err != nil {
		t.Fatal(err)
	}
	assertFile(t, filepath.Join(dst, "a.txt"), "a")
}

func TestRun_AllIgnored(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, ".gitignore"), "*\n")
	writeFile(t, filepath.Join(src, "a.txt"), "a")
	err := Run(Config{Src: src, Dst: t.TempDir(), Commits: 1, Days: 8, Seed: 1, Now: time.Now()})
	if err == nil || !strings.Contains(err.Error(), "no files") {
		t.Fatalf("got %v", err)
	}
}

func TestRun_ReplacesReadmeToday(t *testing.T) {
	requireGit(t)
	src := t.TempDir()
	dst := t.TempDir()
	writeFile(t, filepath.Join(src, "README.md"), "from src")
	writeFile(t, filepath.Join(src, "keep.txt"), "k")
	now := time.Date(2026, 8, 16, 12, 30, 0, 0, time.UTC)
	if err := Run(Config{Src: src, Dst: dst, Commits: 2, Days: 9, Seed: 4, Now: now}); err != nil {
		t.Fatal(err)
	}
	assertFile(t, filepath.Join(dst, "README.md"), "")
	head := gitOutput(t, dst, "show", "--name-only", "--pretty=format:", "HEAD")
	if !strings.Contains(head, "README.md") {
		t.Fatalf("HEAD files %q", head)
	}
}

func TestRun_MessagesOnly(t *testing.T) {
	requireGit(t)
	src := t.TempDir()
	dst := t.TempDir()
	writeFile(t, filepath.Join(src, "a.txt"), "a")
	now := time.Date(2026, 8, 16, 13, 0, 0, 0, time.UTC)
	if err := Run(Config{
		Src:      src,
		Dst:      dst,
		Commits:  2,
		Days:     10,
		Seed:     5,
		Now:      now,
		Messages: []string{"Правки"},
	}); err != nil {
		t.Fatal(err)
	}
	msgs := strings.Split(strings.TrimSpace(gitOutput(t, dst, "log", "--format=%s")), "\n")
	for _, m := range msgs {
		if m != "Правки" {
			t.Fatalf("message %q", m)
		}
	}
	authors := strings.Split(strings.TrimSpace(gitOutput(t, dst, "log", "--format=%an <%ae>")), "\n")
	for _, a := range authors {
		if a != "gitago <gitago@local.lol>" {
			t.Fatalf("author %q", a)
		}
	}
}

func TestRun_CommittersOnly(t *testing.T) {
	requireGit(t)
	src := t.TempDir()
	dst := t.TempDir()
	writeFile(t, filepath.Join(src, "a.txt"), "a")
	now := time.Date(2026, 8, 16, 14, 0, 0, 0, time.UTC)
	who := Identity{Name: "Иван Соколов", Email: "ivan.sokolov@example.com"}
	if err := Run(Config{
		Src:        src,
		Dst:        dst,
		Commits:    2,
		Days:       10,
		Seed:       6,
		Now:        now,
		Committers: []Identity{who},
	}); err != nil {
		t.Fatal(err)
	}
	want := who.Name + " <" + who.Email + ">"
	for _, line := range strings.Split(strings.TrimSpace(gitOutput(t, dst, "log", "--format=%an <%ae>%n%cn <%ce>")), "\n") {
		if line != want {
			t.Fatalf("identity %q want %q", line, want)
		}
	}
	for _, m := range strings.Split(strings.TrimSpace(gitOutput(t, dst, "log", "--format=%s")), "\n") {
		if len(m) < 8 || len(m) > 32 {
			t.Fatalf("expected random message, got %q", m)
		}
	}
}

func TestRun_UsesConfigLists(t *testing.T) {
	requireGit(t)
	src := t.TempDir()
	dst := t.TempDir()
	writeFile(t, filepath.Join(src, "a.txt"), "a")
	now := time.Date(2026, 8, 16, 19, 0, 0, 0, time.UTC)
	pool := []string{"Правки", "Фикс"}
	people := []Identity{
		{Name: "Анна Волкова", Email: "anna.volkova@example.com"},
		{Name: "Иван Соколов", Email: "ivan.sokolov@example.com"},
	}
	if err := Run(Config{
		Src:        src,
		Dst:        dst,
		Commits:    3,
		Days:       12,
		Seed:       11,
		Now:        now,
		Messages:   pool,
		Committers: people,
	}); err != nil {
		t.Fatal(err)
	}

	msgs := strings.Split(strings.TrimSpace(gitOutput(t, dst, "log", "--format=%s")), "\n")
	if len(msgs) != 4 {
		t.Fatalf("messages=%v", msgs)
	}
	allowedMsg := map[string]struct{}{"Правки": {}, "Фикс": {}}
	for _, m := range msgs {
		if _, ok := allowedMsg[m]; !ok {
			t.Fatalf("message %q not from config", m)
		}
	}
	allowedID := map[string]struct{}{
		"Анна Волкова <anna.volkova@example.com>": {},
		"Иван Соколов <ivan.sokolov@example.com>": {},
	}
	ids := strings.Split(strings.TrimSpace(gitOutput(t, dst, "log", "--format=%an <%ae>%n%cn <%ce>")), "\n")
	for _, a := range ids {
		if _, ok := allowedID[a]; !ok {
			t.Fatalf("identity %q", a)
		}
	}
}

func TestRun_DeterministicSeed_WithConfig(t *testing.T) {
	requireGit(t)
	now := time.Date(2026, 8, 16, 16, 0, 0, 0, time.UTC)
	msgs, people, err := LoadYAML(filepath.Join("..", "..", "gitago.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	logOf := func() string {
		src := t.TempDir()
		dst := t.TempDir()
		writeFile(t, filepath.Join(src, "one.txt"), "1")
		if err := Run(Config{
			Src: src, Dst: dst, Commits: 3, Days: 20, Seed: 8, Now: now,
			Messages: msgs, Committers: people,
		}); err != nil {
			t.Fatal(err)
		}
		return gitOutput(t, dst, "log", "--reverse", "--format=%aI %s %an <%ae>")
	}
	if a, b := logOf(), logOf(); a != b {
		t.Fatalf("seed was not deterministic\nA:\n%s\nB:\n%s", a, b)
	}
}

func TestRun_ShippedYAML(t *testing.T) {
	requireGit(t)
	msgs, people, err := LoadYAML(filepath.Join("..", "..", "gitago.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	src := t.TempDir()
	dst := t.TempDir()
	writeFile(t, filepath.Join(src, "a.txt"), "a")
	now := time.Date(2026, 8, 16, 17, 0, 0, 0, time.UTC)
	if err := Run(Config{
		Src: src, Dst: dst, Commits: 4, Days: 30, Seed: 13, Now: now,
		Messages: msgs, Committers: people,
	}); err != nil {
		t.Fatal(err)
	}
	allowedMsg := map[string]struct{}{}
	for _, m := range msgs {
		allowedMsg[m] = struct{}{}
	}
	allowedID := map[string]struct{}{}
	for _, p := range people {
		allowedID[p.Name+" <"+p.Email+">"] = struct{}{}
	}
	for _, m := range strings.Split(strings.TrimSpace(gitOutput(t, dst, "log", "--format=%s")), "\n") {
		if _, ok := allowedMsg[m]; !ok {
			t.Fatalf("message %q not in shipped yaml", m)
		}
	}
	for _, a := range strings.Split(strings.TrimSpace(gitOutput(t, dst, "log", "--format=%an <%ae>")), "\n") {
		if _, ok := allowedID[a]; !ok {
			t.Fatalf("author %q not in shipped yaml", a)
		}
	}
}
