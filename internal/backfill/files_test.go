package backfill

import (
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestCollectFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.txt"), "a")
	writeFile(t, filepath.Join(root, "sub", "b.txt"), "b")
	writeFile(t, filepath.Join(root, ".git", "config"), "nope")
	if err := os.Mkdir(filepath.Join(root, "emptydir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "a.txt"), filepath.Join(root, "link.txt")); err != nil {
		t.Logf("symlink skipped: %v", err)
	}

	got, err := CollectFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.txt", filepath.Join("sub", "b.txt")}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCollectFiles_EmptyAndMissing(t *testing.T) {
	got, err := CollectFiles(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v", got)
	}
	if _, err := CollectFiles(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected error")
	}
}

func TestPartitionFiles_OneGroup(t *testing.T) {
	files := []string{"a", "b", "c"}
	rng := rand.New(rand.NewPCG(1, 1))
	groups := partitionFiles(files, 1, rng)
	if len(groups) != 1 || len(groups[0]) != 3 {
		t.Fatalf("groups=%v", groups)
	}
}

func TestCopyFile_PreservesContentAndTime(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	rel := filepath.Join("nested", "x.bin")
	writeFile(t, filepath.Join(src, rel), "payload")
	if err := os.Chmod(filepath.Join(src, rel), 0o755); err != nil {
		t.Fatal(err)
	}
	when := time.Date(2024, 2, 3, 4, 5, 6, 0, time.Local)
	if err := copyFile(src, dst, rel, when); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dst, rel))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "payload" {
		t.Fatalf("got %q", got)
	}
	info, err := os.Stat(filepath.Join(dst, rel))
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Truncate(time.Second).Equal(when.Truncate(time.Second)) {
		t.Fatalf("mtime %v want %v", info.ModTime(), when)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("expected executable bit, mode %v", info.Mode())
	}
}

func TestPrepareDst(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "a", "b")
	if err := prepareDst(missing); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(missing)
	if err != nil || !st.IsDir() {
		t.Fatalf("stat %v %v", st, err)
	}

	empty := t.TempDir()
	if err := prepareDst(empty); err != nil {
		t.Fatal(err)
	}

	file := filepath.Join(t.TempDir(), "f")
	writeFile(t, file, "x")
	if err := prepareDst(file); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("got %v", err)
	}

	full := t.TempDir()
	writeFile(t, filepath.Join(full, "x"), "x")
	if err := prepareDst(full); err == nil || !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("got %v", err)
	}
}

func TestPartitionFiles_CoversAll(t *testing.T) {
	files := []string{"a", "b", "c", "d", "e"}
	rng := rand.New(rand.NewPCG(1, 2))
	groups := partitionFiles(files, 3, rng)
	if len(groups) != 3 {
		t.Fatalf("groups=%d", len(groups))
	}
	seen := map[string]int{}
	for _, g := range groups {
		for _, f := range g {
			seen[f]++
		}
	}
	if len(seen) != len(files) {
		t.Fatalf("missing files: %v", seen)
	}
	for _, n := range seen {
		if n != 1 {
			t.Fatalf("duplicate assignment: %v", seen)
		}
	}
}

func TestPartitionFiles_MoreCommitsThanFiles(t *testing.T) {
	files := []string{"a", "b"}
	rng := rand.New(rand.NewPCG(3, 4))
	groups := partitionFiles(files, 5, rng)
	if len(groups) != 5 {
		t.Fatalf("groups=%d", len(groups))
	}
	nonEmpty := 0
	for _, g := range groups {
		if len(g) > 0 {
			nonEmpty++
		}
	}
	if nonEmpty != 2 {
		t.Fatalf("nonEmpty=%d", nonEmpty)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
