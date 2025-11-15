package gitrepo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClient_InitAddCommit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	dir := t.TempDir()
	c := Client{}
	if err := c.Init(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.Add(dir, "f.txt"); err != nil {
		t.Fatal(err)
	}
	when := time.Date(2020, 5, 6, 7, 8, 9, 0, time.UTC)
	if err := c.Commit(dir, "abcXYZ12", when, false, "John Smith", "john.smith@mail.lol"); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("git", "-C", dir, "log", "-1", "--format=%aI%n%s%n%an <%ae>")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != 3 {
		t.Fatalf("log=%q", out)
	}
	got, err := time.Parse(time.RFC3339, lines[0])
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(when) {
		t.Fatalf("date %v want %v", got, when)
	}
	if lines[1] != "abcXYZ12" {
		t.Fatalf("message %q", lines[1])
	}
	if lines[2] != "John Smith <john.smith@mail.lol>" {
		t.Fatalf("author %q", lines[2])
	}
}

func TestClient_CommitEmpty(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	c := Client{}
	if err := c.Init(dir); err != nil {
		t.Fatal(err)
	}
	when := time.Date(2021, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := c.Commit(dir, "emptyone", when, true, "", ""); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "-C", dir, "log", "-1", "--format=%s%n%an <%ae>%n%cn <%ce>")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	got := strings.TrimSpace(string(out))
	want := "emptyone\ngitago <gitago@local.lol>\ngitago <gitago@local.lol>"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	branch, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(branch)) != "master" {
		t.Fatalf("branch %q", branch)
	}
}

func TestClient_AddEmpty(t *testing.T) {
	c := Client{}
	if err := c.Add(t.TempDir()); err != nil {
		t.Fatal(err)
	}
}

func TestClient_MissingGit(t *testing.T) {
	c := Client{Bin: filepath.Join(t.TempDir(), "no-git")}
	if err := c.Init(t.TempDir()); err == nil {
		t.Fatal("expected error")
	}
}
