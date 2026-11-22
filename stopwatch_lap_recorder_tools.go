// Package stopwatch_lap_recorder_tools provides a stopwatch that records
// cumulative elapsed time and named laps, and can render a summary table.
package stopwatch_lap_recorder_tools

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Clock abstracts time so tests can supply a deterministic, manual clock.
// The production default is time.Now.
type Clock func() time.Time

// Lap is a named split recorded at a point in time.
//
// Elapsed is the time since the stopwatch was started (or since the last
// Reset) at the moment the lap was recorded. Duration is the time since the
// previous lap, or since the start for the first lap.
type Lap struct {
	Name     string
	Elapsed  time.Duration
	Duration time.Duration
}

// Stopwatch tracks elapsed time and named laps.
//
// The zero value is not usable; construct one with NewStopwatch.
type Stopwatch struct {
	clock   Clock
	start   time.Time
	running bool
	laps    []Lap
}

// NewStopwatch returns a Stopwatch that reads wall-clock time. Pass a custom
// Clock in tests to make behaviour deterministic.
func NewStopwatch() *Stopwatch {
	return NewStopwatchWithClock(time.Now)
}

// NewStopwatchWithClock returns a Stopwatch that reads time from clock.
// It is exported so callers (and tests) can inject a fake time source.
func NewStopwatchWithClock(clock Clock) *Stopwatch {
	if clock == nil {
		clock = time.Now
	}
	return &Stopwatch{clock: clock}
}

// Start begins timing. It is safe to call Start multiple times; only the
// first call after a Reset (or on a fresh Stopwatch) takes effect. Later
// calls while running are no-ops, which avoids accidentally truncating
// elapsed time.
func (s *Stopwatch) Start() {
	if s.running {
		return
	}
	s.start = s.clock()
	s.running = true
}

// Elapsed returns the time since the stopwatch was started. If the
// stopwatch has never been started, it returns zero.
func (s *Stopwatch) Elapsed() time.Duration {
	if !s.running {
		return 0
	}
	return s.clock().Sub(s.start)
}

// Lap records a lap named name. The name may be empty; empty names are
// rendered as "lap N" in the summary table. Recording a lap before Start
// is a no-op, since there is no elapsed time to measure against.
func (s *Stopwatch) Lap(name string) {
	if !s.running {
		return
	}
	elapsed := s.clock().Sub(s.start)
	var sinceLast time.Duration
	if len(s.laps) == 0 {
		sinceLast = elapsed
	} else {
		sinceLast = elapsed - s.laps[len(s.laps)-1].Elapsed
	}
	s.laps = append(s.laps, Lap{Name: name, Elapsed: elapsed, Duration: sinceLast})
}

// Laps returns a copy of the recorded laps in the order they were recorded.
// Returning a copy keeps the internal slice from being mutated by callers.
func (s *Stopwatch) Laps() []Lap {
	out := make([]Lap, len(s.laps))
	copy(out, s.laps)
	return out
}

// Reset clears all laps and stops the stopwatch. A subsequent Start begins a
// fresh measurement.
func (s *Stopwatch) Reset() {
	s.start = time.Time{}
	s.running = false
	s.laps = nil
}

// Summary renders the recorded laps as a fixed-width table.
//
// Columns are: lap name, cumulative elapsed time, and duration since the
// previous lap. Durations are formatted with millisecond precision, which is
// enough for human-readable lap summaries and keeps the table narrow.
func (s *Stopwatch) Summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-12s %-14s %-14s\n", "LAP", "ELAPSED", "DURATION")
	for i, lap := range s.laps {
		name := lap.Name
		if name == "" {
			name = fmt.Sprintf("lap %d", i+1)
		}
		fmt.Fprintf(&b, "%-12s %-14s %-14s\n", name, formatDuration(lap.Elapsed), formatDuration(lap.Duration))
	}
	return b.String()
}

// SortedLaps returns the laps ordered by cumulative elapsed time, ascending.
// Ties are broken by the original recording order so the result is stable.
// This is useful when laps were recorded out of order by an external source.
func (s *Stopwatch) SortedLaps() []Lap {
	out := s.Laps()
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Elapsed < out[j].Elapsed
	})
	return out
}

// formatDuration renders a duration as seconds with three decimal places.
func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.3fs", d.Seconds())
}
