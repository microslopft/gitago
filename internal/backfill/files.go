package backfill

import (
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

// CollectFiles returns relative paths of regular files under root.
// Nested .git directories are skipped. Paths matched by any .gitignore
// (root or nested) are omitted, including files under ignored directories.
func CollectFiles(root string) ([]string, error) {
	ign := newTreeIgnore()
	var files []string
	if err := collectWalk(root, "", ign, &files); err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func collectWalk(root, rel string, ign *treeIgnore, files *[]string) error {
	dirpath := root
	if rel != "" {
		dirpath = filepath.Join(root, rel)
	}
	entries, err := os.ReadDir(dirpath)
	if err != nil {
		return err
	}

	relSlash := filepath.ToSlash(rel)
	for _, e := range entries {
		if e.Name() != ".gitignore" || !e.Type().IsRegular() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dirpath, e.Name()))
		if err != nil {
			return err
		}
		rules, err := parseGitignore(relSlash, string(data))
		if err != nil {
			return fmt.Errorf("parse %s: %w", filepath.Join(rel, ".gitignore"), err)
		}
		ign.add(relSlash, rules)
	}

	for _, e := range entries {
		child := e.Name()
		if rel != "" {
			child = filepath.Join(rel, e.Name())
		}
		if e.IsDir() {
			if e.Name() == ".git" {
				continue
			}
			if ign.ignored(child, true) {
				continue
			}
			if err := collectWalk(root, child, ign, files); err != nil {
				return err
			}
			continue
		}
		if !e.Type().IsRegular() {
			continue
		}
		if ign.ignored(child, false) {
			continue
		}
		*files = append(*files, child)
	}
	return nil
}

func partitionFiles(files []string, n int, rng *rand.Rand) [][]string {
	shuffled := append([]string(nil), files...)
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	groups := make([][]string, n)
	for i, name := range shuffled {
		g := i % n
		groups[g] = append(groups[g], name)
	}
	return groups
}

func copyFile(srcRoot, dstRoot, rel string, when time.Time) error {
	srcPath := filepath.Join(srcRoot, rel)
	dstPath := filepath.Join(dstRoot, rel)
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}

	in, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return setTimes(dstPath, when)
}

func setTimes(path string, when time.Time) error {
	if err := os.Chtimes(path, when, when); err != nil {
		return fmt.Errorf("chtimes %s: %w", path, err)
	}
	if runtime.GOOS == "darwin" {
		stamp := when.Format("01/02/2006 15:04:05")
		// SetFile is optional (needs Developer Tools). Creation date is best-effort.
		_ = runQuiet("SetFile", "-d", stamp, "-m", stamp, path)
	}
	return nil
}

func prepareDst(dst string) error {
	st, err := os.Stat(dst)
	if os.IsNotExist(err) {
		return os.MkdirAll(dst, 0o755)
	}
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("dst %q is not a directory", dst)
	}
	entries, err := os.ReadDir(dst)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("dst %q is not empty", dst)
	}
	return nil
}
