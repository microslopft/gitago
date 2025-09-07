package backfill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConfigNormalize(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	writeFile(t, filepath.Join(src, "x.txt"), "x")

	srcFile := filepath.Join(t.TempDir(), "notdir")
	writeFile(t, srcFile, "nope")

	nestedDst := filepath.Join(src, "out")
	if err := os.Mkdir(nestedDst, 0o755); err != nil {
		t.Fatal(err)
	}
	nestedSrc := filepath.Join(dst, "in")
	if err := os.Mkdir(nestedSrc, 0o755); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
		after   func(t *testing.T, cfg Config)
	}{
		{
			name: "ok fills abs and now",
			cfg:  Config{Src: src, Dst: dst, Commits: 1, Days: 10},
			after: func(t *testing.T, cfg Config) {
				if !filepath.IsAbs(cfg.Src) || !filepath.IsAbs(cfg.Dst) {
					t.Fatalf("expected abs paths: %+v", cfg)
				}
				if cfg.Now.IsZero() {
					t.Fatal("Now should be filled")
				}
			},
		},
		{
			name: "preserves Now",
			cfg:  Config{Src: src, Dst: dst, Commits: 2, Days: 3, Now: now},
			after: func(t *testing.T, cfg Config) {
				if !cfg.Now.Equal(now) {
					t.Fatalf("Now=%v", cfg.Now)
				}
			},
		},
		{name: "empty src", cfg: Config{Dst: dst, Commits: 1, Days: 1}, wantErr: "src is required"},
		{name: "empty dst", cfg: Config{Src: src, Commits: 1, Days: 1}, wantErr: "dst is required"},
		{name: "commits zero", cfg: Config{Src: src, Dst: dst, Commits: 0, Days: 1}, wantErr: "commits"},
		{name: "days zero", cfg: Config{Src: src, Dst: dst, Commits: 1, Days: 0}, wantErr: "days"},
		{name: "src missing", cfg: Config{Src: filepath.Join(src, "nope"), Dst: dst, Commits: 1, Days: 1}, wantErr: "src"},
		{name: "src not dir", cfg: Config{Src: srcFile, Dst: dst, Commits: 1, Days: 1}, wantErr: "not a directory"},
		{name: "same dir", cfg: Config{Src: src, Dst: src, Commits: 1, Days: 1}, wantErr: "distinct"},
		{name: "dst inside src", cfg: Config{Src: src, Dst: nestedDst, Commits: 1, Days: 1}, wantErr: "distinct"},
		{name: "src inside dst", cfg: Config{Src: nestedSrc, Dst: dst, Commits: 1, Days: 1}, wantErr: "distinct"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.cfg
			err := cfg.normalize()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				if tc.after != nil {
					tc.after(t, cfg)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err=%v want substring %q", err, tc.wantErr)
			}
		})
	}
}
