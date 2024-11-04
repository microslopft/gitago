package backfill

import (
	"math/rand/v2"
	"testing"
	"time"
	"unicode"
)

func TestRandomCommitTimes(t *testing.T) {
	rng := rand.New(rand.NewPCG(11, 22))
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 16, 15, 0, 0, 0, time.UTC)
	times, err := randomCommitTimes(rng, start, end, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(times) != 20 {
		t.Fatalf("len=%d", len(times))
	}
	seen := map[int64]struct{}{}
	for i, tm := range times {
		if tm.Before(start) || !tm.Before(end) {
			t.Fatalf("time %v out of range", tm)
		}
		if i > 0 && !times[i].After(times[i-1]) {
			t.Fatalf("not strictly increasing: %v then %v", times[i-1], tm)
		}
		seen[tm.Unix()] = struct{}{}
	}
	if len(seen) != 20 {
		t.Fatalf("duplicate timestamps: %d unique", len(seen))
	}
}

func TestRandomCommitTimes_WindowTooSmall(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	start := time.Unix(0, 0).UTC()
	end := start.Add(3 * time.Second)
	_, err := randomCommitTimes(rng, start, end, 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRandomMessage(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 6))
	seen := map[string]struct{}{}
	for i := 0; i < 20; i++ {
		msg := randomMessage(rng)
		if len(msg) < 8 || len(msg) > 32 {
			t.Fatalf("len=%d msg=%q", len(msg), msg)
		}
		for _, r := range msg {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				t.Fatalf("unexpected rune %q in %q", r, msg)
			}
		}
		seen[msg] = struct{}{}
	}
	if len(seen) < 2 {
		t.Fatal("messages are not random")
	}
}

func TestPickMessage_FromList(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	pool := []string{"Правки", "Фикс"}
	seen := map[string]struct{}{}
	for i := 0; i < 40; i++ {
		got := pickMessage(rng, pool)
		if got != "Правки" && got != "Фикс" {
			t.Fatalf("unexpected %q", got)
		}
		seen[got] = struct{}{}
	}
	if len(seen) != 2 {
		t.Fatalf("expected both messages, got %v", seen)
	}
}

func TestPickMessage_EmptyFallsBack(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	got := pickMessage(rng, nil)
	if len(got) < 8 {
		t.Fatalf("fallback %q", got)
	}
}

func TestPickCommitter(t *testing.T) {
	rng := rand.New(rand.NewPCG(9, 8))
	if got := pickCommitter(rng, nil); got != (Identity{}) {
		t.Fatalf("empty list: %+v", got)
	}
	people := []Identity{
		{Name: "Анна Волкова", Email: "anna@example.com"},
		{Name: "Иван Соколов", Email: "john.smith@mail.lol"},
	}
	got := pickCommitter(rng, people)
	if got != people[0] && got != people[1] {
		t.Fatalf("got %+v", got)
	}
}

func TestRandomCommitTimes_NonPositiveWindow(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	start := time.Unix(10, 0).UTC()
	_, err := randomCommitTimes(rng, start, start, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRandomCommitTimes_One(t *testing.T) {
	rng := rand.New(rand.NewPCG(2, 3))
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(10 * time.Second)
	times, err := randomCommitTimes(rng, start, end, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(times) != 1 || times[0].Before(start) || !times[0].Before(end) {
		t.Fatalf("times=%v", times)
	}
}

func TestRandomTimeBetween_EqualTimes(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	tm := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if got := randomTimeBetween(rng, tm, tm); !got.Equal(tm) {
		t.Fatalf("got %v", got)
	}
}

func TestRandomTimeBetween(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 8))
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 50; i++ {
		got := randomTimeBetween(rng, start, end)
		if got.Before(start) || got.After(end) {
			t.Fatalf("out of range: %v", got)
		}
	}
}
