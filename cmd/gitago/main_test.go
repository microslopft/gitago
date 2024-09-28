package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	cfg, err := parseArgs([]string{"-src", "/tmp/src", "-dst", "/tmp/dst", "-commits", "7", "-days", "30", "-seed", "99"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Src != "/tmp/src" || cfg.Dst != "/tmp/dst" || cfg.Commits != 7 || cfg.Days != 30 || cfg.Seed != 99 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if len(cfg.Messages) != 0 || len(cfg.Committers) != 0 {
		t.Fatalf("lists should be empty: %+v", cfg)
	}
}

func TestParseArgs_Defaults(t *testing.T) {
	cfg, err := parseArgs([]string{"-src", "/tmp/src", "-dst", "/tmp/dst", "-commits", "1"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Days != 365 || cfg.Seed != 0 {
		t.Fatalf("defaults: days=%d seed=%d", cfg.Days, cfg.Seed)
	}
}

func TestParseArgs_Config(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gitago.yaml")
	if err := os.WriteFile(path, []byte("messages:\n  - Правки\ncommitters:\n  - name: Иван\n    email: john.smith@mail.lol\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := parseArgs([]string{"-src", "/tmp/src", "-dst", "/tmp/dst", "-commits", "2", "-config", path}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Messages) != 1 || cfg.Messages[0] != "Правки" {
		t.Fatalf("messages=%v", cfg.Messages)
	}
	if len(cfg.Committers) != 1 || cfg.Committers[0].Email != "john.smith@mail.lol" {
		t.Fatalf("committers=%v", cfg.Committers)
	}
}

func TestParseArgs_Errors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing dst", args: []string{"-src", "/tmp/src"}, want: "required"},
		{name: "missing src", args: []string{"-dst", "/tmp/dst", "-commits", "1"}, want: "required"},
		{name: "commits zero", args: []string{"-src", "/a", "-dst", "/b", "-commits", "0"}, want: "required"},
		{name: "unknown flag", args: []string{"-nope"}, want: "flag provided but not defined"},
		{name: "missing config", args: []string{"-src", "/a", "-dst", "/b", "-commits", "1", "-config", filepath.Join(t.TempDir(), "nope.yaml")}, want: "read config"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseArgs(tc.args, io.Discard)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err=%v want %q", err, tc.want)
			}
		})
	}
}

func TestParseArgs_Help(t *testing.T) {
	var buf bytes.Buffer
	_, err := parseArgs([]string{"-h"}, &buf)
	if err == nil {
		t.Fatal("expected flag.ErrHelp")
	}
	if err != flag.ErrHelp {
		t.Fatalf("err=%v", err)
	}
	out := buf.String()
	for _, flagName := range []string{"-src", "-dst", "-commits", "-days", "-seed", "-config"} {
		if !strings.Contains(out, flagName) {
			t.Fatalf("usage missing %s:\n%s", flagName, out)
		}
	}
}

func TestRun_EndToEnd(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "out")
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(t.TempDir(), "gitago.yaml")
	if err := os.WriteFile(cfgPath, []byte("messages:\n  - Правки\ncommitters:\n  - name: Анна\n    email: anna@example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{
		"-src", src, "-dst", dst, "-commits", "1", "-days", "10", "-seed", "1", "-config", cfgPath,
	}, io.Discard); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("git", "-C", dst, "log", "--format=%s %an <%ae>").CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	got := strings.TrimSpace(string(out))
	for _, line := range strings.Split(got, "\n") {
		if line != "Правки Анна <anna@example.com>" {
			t.Fatalf("log line %q\nfull:\n%s", line, got)
		}
	}
	body, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "a" {
		t.Fatalf("copied %q", body)
	}
}
