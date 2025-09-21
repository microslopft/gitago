package backfill

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"time"

	"github.com/microslopft/gitago/internal/gitrepo"
)

const readmeName = "README.md"

// Run copies non-ignored files from src into dst across Commits backdated
// git commits, then adds an empty README.md dated Now.
func Run(cfg Config) error {
	if err := cfg.normalize(); err != nil {
		return err
	}

	files, err := CollectFiles(cfg.Src)
	if err != nil {
		return fmt.Errorf("scan src: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no files in src %q", cfg.Src)
	}

	if err := prepareDst(cfg.Dst); err != nil {
		return err
	}

	git := cfg.Git
	if git == nil {
		git = gitrepo.Client{}
	}
	if err := git.Init(cfg.Dst); err != nil {
		return err
	}

	seed := uint64(cfg.Seed)
	if cfg.Seed == 0 {
		seed = uint64(time.Now().UnixNano())
	}
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))

	start := cfg.Now.AddDate(0, 0, -cfg.Days)
	times, err := randomCommitTimes(rng, start, cfg.Now, cfg.Commits)
	if err != nil {
		return err
	}
	groups := partitionFiles(files, cfg.Commits, rng)

	for i := 0; i < cfg.Commits; i++ {
		when := times[i]
		for _, rel := range groups[i] {
			mtime := randomTimeBetween(rng, start, when)
			if err := copyFile(cfg.Src, cfg.Dst, rel, mtime); err != nil {
				return fmt.Errorf("copy %s: %w", rel, err)
			}
		}
		msg := pickMessage(rng, cfg.Messages)
		who := pickCommitter(rng, cfg.Committers)
		if len(groups[i]) == 0 {
			if err := git.Commit(cfg.Dst, msg, when, true, who.Name, who.Email); err != nil {
				return err
			}
			continue
		}
		if err := git.Add(cfg.Dst, groups[i]...); err != nil {
			return err
		}
		if err := git.Commit(cfg.Dst, msg, when, false, who.Name, who.Email); err != nil {
			return err
		}
	}

	readme := filepath.Join(cfg.Dst, readmeName)
	if err := os.WriteFile(readme, nil, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", readmeName, err)
	}
	if err := setTimes(readme, cfg.Now); err != nil {
		return err
	}
	if err := git.Add(cfg.Dst, readmeName); err != nil {
		return err
	}

	who := pickCommitter(rng, cfg.Committers)
	return git.Commit(cfg.Dst, pickMessage(rng, cfg.Messages), cfg.Now, false, who.Name, who.Email)
}
