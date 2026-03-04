package backfill

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCollectFiles_Gitignore(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".gitignore"), ""+
		"# comment\n"+
		"*.log\n"+
		"!keep.log\n"+
		"/secret.txt\n"+
		"build/\n"+
		"skip/\n"+
		"!skip/keep.txt\n"+
		"doc/*.txt\n"+
		"**/tmpbin\n")
	writeFile(t, filepath.Join(root, "a.txt"), "a")
	writeFile(t, filepath.Join(root, "a.log"), "ignored")
	writeFile(t, filepath.Join(root, "keep.log"), "kept")
	writeFile(t, filepath.Join(root, "secret.txt"), "root secret")
	writeFile(t, filepath.Join(root, "other", "secret.txt"), "nested secret")
	writeFile(t, filepath.Join(root, "build", "out.go"), "nope")
	writeFile(t, filepath.Join(root, "skip", "keep.txt"), "still ignored via parent")
	writeFile(t, filepath.Join(root, "doc", "n.txt"), "ignored")
	writeFile(t, filepath.Join(root, "doc", "sub", "n.txt"), "kept")
	writeFile(t, filepath.Join(root, "deep", "tmpbin"), "ignored")
	writeFile(t, filepath.Join(root, "sub", ".gitignore"), "*.tmp\n!keep.tmp\n")
	writeFile(t, filepath.Join(root, "sub", "c.txt"), "c")
	writeFile(t, filepath.Join(root, "sub", "d.tmp"), "ignored")
	writeFile(t, filepath.Join(root, "sub", "keep.tmp"), "kept")
	writeFile(t, filepath.Join(root, "sub", "nested", "e.tmp"), "ignored")
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, ".git", "config"), "nope")

	got, err := CollectFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		".gitignore",
		"a.txt",
		filepath.Join("doc", "sub", "n.txt"),
		"keep.log",
		filepath.Join("other", "secret.txt"),
		filepath.Join("sub", ".gitignore"),
		filepath.Join("sub", "c.txt"),
		filepath.Join("sub", "keep.tmp"),
	}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

func TestCompileRule_Match(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		isDir   bool
		want    bool
	}{
		{"*.log", "a.log", false, true},
		{"*.log", "dir/a.log", false, true},
		{"*.log", "a.txt", false, false},
		{"/secret.txt", "secret.txt", false, true},
		{"/secret.txt", "other/secret.txt", false, false},
		{"build/", "build", true, true},
		{"build/", "lib/build", true, true},
		{"build/", "build", false, false},
		{"doc/*.txt", "doc/n.txt", false, true},
		{"doc/*.txt", "doc/sub/n.txt", false, false},
		{"**/tmpbin", "tmpbin", false, true},
		{"**/tmpbin", "deep/tmpbin", false, true},
		{"foo/**", "foo", true, false},
		{"foo/**", "foo/bar", false, true},
		{"a/**/b", "a/b", false, true},
		{"a/**/b", "a/x/b", false, true},
		{"a/**/b", "a/x/y/b", false, true},
		{"a/**/b", "z/b", false, false},
		{"[ab].txt", "a.txt", false, true},
		{"[ab].txt", "c.txt", false, false},
		{"?.txt", "a.txt", false, true},
		{"?.txt", "ab.txt", false, false},
		{`\!keep`, "!keep", false, true},
		{"foo/**", "foo/a/b", false, true},
		{"**", "anything/here", false, true},
		{"unclosed[.txt", "unclosed[.txt", false, true},
	}
	for _, tc := range tests {
		rule, err := compileRule("", tc.pattern)
		if err != nil {
			t.Fatalf("pattern %q: %v", tc.pattern, err)
		}
		got := rule.matches(tc.path, tc.isDir)
		if got != tc.want {
			t.Errorf("pattern %q path %q dir=%v: got %v want %v", tc.pattern, tc.path, tc.isDir, got, tc.want)
		}
	}
}

func TestGitignore_NegationAndParentDir(t *testing.T) {
	ign := newTreeIgnore()
	rules, err := parseGitignore("", "*.log\n!keep.log\nskip/\n!skip/keep.txt\n")
	if err != nil {
		t.Fatal(err)
	}
	ign.add("", rules)
	if ign.ignored("keep.log", false) {
		t.Fatal("keep.log should be re-included")
	}
	if !ign.ignored("a.log", false) {
		t.Fatal("a.log should be ignored")
	}
	if !ign.ignored("skip", true) {
		t.Fatal("skip/ should be ignored")
	}
	if !ign.ignored("skip/keep.txt", false) {
		t.Fatal("cannot re-include a file under an ignored directory")
	}
}

func TestGitignore_NestedOverrides(t *testing.T) {
	ign := newTreeIgnore()
	rootRules, err := parseGitignore("", "*.tmp\n")
	if err != nil {
		t.Fatal(err)
	}
	ign.add("", rootRules)
	subRules, err := parseGitignore("sub", "!keep.tmp\n")
	if err != nil {
		t.Fatal(err)
	}
	ign.add("sub", subRules)
	if !ign.ignored("a.tmp", false) {
		t.Fatal("root *.tmp")
	}
	if ign.ignored("sub/keep.tmp", false) {
		t.Fatal("nested negation should re-include")
	}
	if !ign.ignored("sub/other.tmp", false) {
		t.Fatal("other tmp still ignored")
	}
}

func TestParseGitignore_SkipsCommentsAndBlanks(t *testing.T) {
	rules, err := parseGitignore("", "\ufeff# hi\n\n*.log\n  \n")
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 {
		t.Fatalf("rules=%d", len(rules))
	}
}

func TestParseGitignore_EscapedTrailingSpaceAndCR(t *testing.T) {
	rules, err := parseGitignore("", "foo\\ \r\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 {
		t.Fatalf("rules=%d", len(rules))
	}
	if !rules[0].matches("foo ", false) {
		t.Fatal("expected match of trailing space name")
	}
}

func TestIgnored_RootAndDot(t *testing.T) {
	ign := newTreeIgnore()
	if ign.ignored("", false) || ign.ignored(".", true) {
		t.Fatal("root must not be ignored")
	}
	ign.add(".", nil)
}
