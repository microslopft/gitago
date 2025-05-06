package backfill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gitago.yaml")
	body := "" +
		"messages:\n" +
		"  - Правки\n" +
		"  - \"  Фикс  \"\n" +
		"  - \"\"\n" +
		"committers:\n" +
		"  - name: John Smith\n" +
		"    email: john.smith@mail.lol\n" +
		"  - name: \"  Анна  \"\n" +
		"    email: \"  anna@example.com  \"\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	msgs, people, err := LoadYAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 || msgs[0] != "Правки" || msgs[1] != "Фикс" {
		t.Fatalf("messages=%v", msgs)
	}
	if len(people) != 2 {
		t.Fatalf("committers=%v", people)
	}
	if people[0] != (Identity{Name: "John Smith", Email: "john.smith@mail.lol"}) {
		t.Fatalf("first=%+v", people[0])
	}
	if people[1] != (Identity{Name: "Анна", Email: "anna@example.com"}) {
		t.Fatalf("second=%+v", people[1])
	}
}

func TestLoadYAML_Errors(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	tests := []struct {
		name    string
		path    string
		body    string
		wantErr string
	}{
		{name: "missing", path: filepath.Join(dir, "nope.yaml"), wantErr: "read config"},
		{name: "invalid yaml", body: "messages: [", wantErr: "parse"},
		{name: "multiline message", body: "messages:\n  - |\n    two\n    lines\n", wantErr: "single line"},
		{name: "email with space", body: "committers:\n  - name: Иван\n    email: ivan @example.com\n", wantErr: "email is invalid"},
		{name: "name only", body: "committers:\n  - name: Иван\n", wantErr: "name and email"},
		{name: "email only", body: "committers:\n  - email: john.smith@mail.lol\n", wantErr: "name and email"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.path
			if path == "" {
				path = write(tc.name+".yaml", tc.body)
			}
			_, _, err := LoadYAML(path)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err=%v want %q", err, tc.wantErr)
			}
		})
	}
}

func TestLoadYAML_EmptyAndPartial(t *testing.T) {
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(empty, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	msgs, people, err := LoadYAML(empty)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 0 || len(people) != 0 {
		t.Fatalf("msgs=%v people=%v", msgs, people)
	}

	onlyMsg := filepath.Join(dir, "msg.yaml")
	if err := os.WriteFile(onlyMsg, []byte("messages:\n  - Правки\ncommitters:\n  - name: \"\"\n    email: \"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	msgs, people, err = LoadYAML(onlyMsg)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 || msgs[0] != "Правки" || len(people) != 0 {
		t.Fatalf("msgs=%v people=%v", msgs, people)
	}
}

func TestShippedYAML(t *testing.T) {
	path := filepath.Join("..", "..", "gitago.yaml")
	msgs, people, err := LoadYAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) < 20 {
		t.Fatalf("too few messages: %d", len(msgs))
	}
	if len(people) < 1 {
		t.Fatalf("too few committers: %d", len(people))
	}
	seenMsg := map[string]struct{}{}
	for _, m := range msgs {
		if m == "" {
			t.Fatal("empty message")
		}
		if _, ok := seenMsg[m]; ok {
			t.Fatalf("duplicate message %q", m)
		}
		seenMsg[m] = struct{}{}
	}
	seenMail := map[string]struct{}{}
	for _, p := range people {
		if p.Name == "" || p.Email == "" {
			t.Fatalf("incomplete %+v", p)
		}
		if _, ok := seenMail[p.Email]; ok {
			t.Fatalf("duplicate email %s", p.Email)
		}
		if !strings.Contains(p.Email, "@") {
			t.Fatalf("email %q", p.Email)
		}
		seenMail[p.Email] = struct{}{}
	}
	for _, want := range []string{"chore: initial commit", "fix: tweaks", "fix: hotfix"} {
		if _, ok := seenMsg[want]; !ok {
			t.Fatalf("shipped yaml missing %q", want)
		}
	}
}
