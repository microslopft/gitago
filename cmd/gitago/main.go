package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/microslopft/gitago/internal/backfill"
)

const usage = `gitago copies files from a source tree into a new git repository,
spreading them across backdated commits, then finishes today with an empty README.md.

Usage:
  gitago -src <dir> -dst <dir> -commits <n> [options]

Flags:
`

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "gitago: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stderr io.Writer) error {
	cfg, err := parseArgs(args, stderr)
	if err != nil {
		return err
	}
	return backfill.Run(cfg)
}

func parseArgs(args []string, stderr io.Writer) (backfill.Config, error) {
	fs := flag.NewFlagSet("gitago", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprint(stderr, usage)
		fs.PrintDefaults()
	}

	var cfg backfill.Config
	var configPath string
	fs.StringVar(&cfg.Src, "src", "", "source directory with files to copy")
	fs.StringVar(&cfg.Dst, "dst", "", "empty destination directory for the new git repo")
	fs.IntVar(&cfg.Commits, "commits", 0, "number of backdated commits that receive source files")
	fs.IntVar(&cfg.Days, "days", 365, "how far back the first commit may be, in days")
	fs.Int64Var(&cfg.Seed, "seed", 0, "RNG seed; 0 uses the current time")
	fs.StringVar(&configPath, "config", "", "optional YAML with commit messages and committers")
	if err := fs.Parse(args); err != nil {
		return backfill.Config{}, err
	}
	if cfg.Src == "" || cfg.Dst == "" || cfg.Commits < 1 {
		fs.Usage()
		return backfill.Config{}, fmt.Errorf("-src, -dst and -commits (>= 1) are required")
	}
	if configPath != "" {
		messages, committers, err := backfill.LoadYAML(configPath)
		if err != nil {
			return backfill.Config{}, err
		}
		cfg.Messages = messages
		cfg.Committers = committers
	}
	return cfg, nil
}
