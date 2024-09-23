package backfill

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"time"
)

const messageCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomCommitTimes(rng *rand.Rand, start, end time.Time, n int) ([]time.Time, error) {
	if !end.After(start) {
		return nil, fmt.Errorf("time window must be positive")
	}
	seconds := int64(end.Sub(start) / time.Second)
	if seconds < int64(n) {
		return nil, fmt.Errorf("time window (%d seconds) is too small for %d commits", seconds, n)
	}

	picked := pickUnique(rng, seconds, n)
	times := make([]time.Time, n)
	for i, sec := range picked {
		times[i] = start.Add(time.Duration(sec) * time.Second)
	}
	slices.SortFunc(times, func(a, b time.Time) int {
		return a.Compare(b)
	})
	return times, nil
}

func pickUnique(rng *rand.Rand, max int64, n int) []int64 {
	used := make(map[int64]struct{}, n)
	out := make([]int64, 0, n)
	for len(out) < n {
		v := rng.Int64N(max)
		if _, ok := used[v]; ok {
			continue
		}
		used[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func randomTimeBetween(rng *rand.Rand, start, end time.Time) time.Time {
	if !end.After(start) {
		return start
	}
	delta := int64(end.Sub(start))
	return start.Add(time.Duration(rng.Int64N(delta + 1)))
}

func randomMessage(rng *rand.Rand) string {
	n := 8 + rng.IntN(25)
	b := make([]byte, n)
	for i := range b {
		b[i] = messageCharset[rng.IntN(len(messageCharset))]
	}
	return string(b)
}

func pickMessage(rng *rand.Rand, messages []string) string {
	if len(messages) == 0 {
		return randomMessage(rng)
	}
	return messages[rng.IntN(len(messages))]
}

func pickCommitter(rng *rand.Rand, people []Identity) Identity {
	if len(people) == 0 {
		return Identity{}
	}
	return people[rng.IntN(len(people))]
}
