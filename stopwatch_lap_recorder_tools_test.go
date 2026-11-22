package stopwatch_lap_recorder_tools

import (
	"strings"
	"testing"
	"time"
)

// fakeClock is a manually advanced clock for deterministic tests.
type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time { return c.now }
func (c *fakeClock) Advance(d time.Duration) {
	c.now = c.now.Add(d)
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
}

func TestStartIsIdempotentWhileRunning(t *testing.T) {
	clk := newFakeClock()
	sw := NewStopwatchWithClock(clk.Now)
	sw.Start()
	clk.Advance(5 * time.Second)
	sw.Start() // should not reset the start time
	clk.Advance(3 * time.Second)

	got := sw.Elapsed()
	want := 8 * time.Second
	if got != want {
		t.Fatalf("Elapsed() = %v, want %v", got, want)
	}
}

func TestElapsedBeforeStartIsZero(t *testing.T) {
	clk := newFakeClock()
	sw := NewStopwatchWithClock(clk.Now)
	clk.Advance(10 * time.Second)
	if got := sw.Elapsed(); got != 0 {
		t.Fatalf("Elapsed() before Start = %v, want 0", got)
	}
}

func TestLapBeforeStartIsNoop(t *testing.T) {
	clk := newFakeClock()
	sw := NewStopwatchWithClock(clk.Now)
	sw.Lap("too early")
	if got := len(sw.Laps()); got != 0 {
		t.Fatalf("Laps() length = %d, want 0", got)
	}
}

func TestLapDurations(t *testing.T) {
	clk := newFakeClock()
	sw := NewStopwatchWithClock(clk.Now)
	sw.Start()
	clk.Advance(2 * time.Second)
	sw.Lap("first")
	clk.Advance(3 * time.Second)
	sw.Lap("second")
	clk.Advance(4 * time.Second)
	sw.Lap("third")

	laps := sw.Laps()
	if len(laps) != 3 {
		t.Fatalf("len(laps) = %d, want 3", len(laps))
	}

	wantElapsed := []time.Duration{2 * time.Second, 5 * time.Second, 9 * time.Second}
	wantDuration := []time.Duration{2 * time.Second, 3 * time.Second, 4 * time.Second}
	for i, lap := range laps {
		if lap.Elapsed != wantElapsed[i] {
			t.Errorf("lap %d Elapsed = %v, want %v", i, lap.Elapsed, wantElapsed[i])
		}
		if lap.Duration != wantDuration[i] {
			t.Errorf("lap %d Duration = %v, want %v", i, lap.Duration, wantDuration[i])
		}
	}
}

func TestLapsReturnsCopy(t *testing.T) {
	clk := newFakeClock()
	sw := NewStopwatchWithClock(clk.Now)
	sw.Start()
	clk.Advance(time.Second)
	sw.Lap("a")

	laps := sw.Laps()
	laps[0].Name = "mutated"

	again := sw.Laps()
	if again[0].Name != "a" {
		t.Fatalf("internal laps mutated through returned slice: %q", again[0].Name)
	}
}

func TestResetClearsState(t *testing.T) {
	clk := newFakeClock()
	sw := NewStopwatchWithClock(clk.Now)
	sw.Start()
	clk.Advance(time.Second)
	sw.Lap("a")
	sw.Reset()

	if got := len(sw.Laps()); got != 0 {
		t.Fatalf("len(Laps()) after Reset = %d, want 0", got)
	}
	if got := sw.Elapsed(); got != 0 {
		t.Fatalf("Elapsed() after Reset = %v, want 0", got)
	}

	// A fresh Start after Reset works.
	sw.Start()
	clk.Advance(2 * time.Second)
	if got := sw.Elapsed(); got != 2*time.Second {
		t.Fatalf("Elapsed() after restart = %v, want 2s", got)
	}
}

func TestSummaryFormatsNamedAndUnnamedLaps(t *testing.T) {
	clk := newFakeClock()
	sw := NewStopwatchWithClock(clk.Now)
	sw.Start()
	clk.Advance(1500 * time.Millisecond)
	sw.Lap("warmup")
	clk.Advance(2500 * time.Millisecond)
	sw.Lap("")

	got := sw.Summary()

	for _, want := range []string{
		"LAP",
		"ELAPSED",
		"DURATION",
		"warmup",
		"1.500s",
		"4.000s",
		"2.500s",
		"lap 2",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Summary() missing %q; got:\n%s", want, got)
		}
	}
}

func TestSortedLapsOrdersByElapsed(t *testing.T) {
	clk := newFakeClock()
	sw := NewStopwatchWithClock(clk.Now)
	sw.Start()
	clk.Advance(3 * time.Second)
	sw.Lap("c")
	clk.Advance(1 * time.Second)
	sw.Lap("a")
	clk.Advance(2 * time.Second)
	sw.Lap("b")

	sorted := sw.SortedLaps()
	wantNames := []string{"c", "a", "b"}
	for i, lap := range sorted {
		if lap.Name != wantNames[i] {
			t.Errorf("sorted[%d].Name = %q, want %q", i, lap.Name, wantNames[i])
		}
	}

	// Original order must be untouched.
	orig := sw.Laps()
	wantOrig := []string{"c", "a", "b"}
	for i, lap := range orig {
		if lap.Name != wantOrig[i] {
			t.Errorf("original[%d].Name = %q, want %q", i, lap.Name, wantOrig[i])
		}
	}
}

func TestNewStopwatchUsesWallClock(t *testing.T) {
	sw := NewStopwatch()
	sw.Start()
	if got := sw.Elapsed(); got < 0 {
		t.Fatalf("Elapsed() = %v, want non-negative", got)
	}
}

func TestNilClockFallsBackToTimeNow(t *testing.T) {
	sw := NewStopwatchWithClock(nil)
	sw.Start()
	if got := sw.Elapsed(); got < 0 {
		t.Fatalf("Elapsed() = %v, want non-negative", got)
	}
}
