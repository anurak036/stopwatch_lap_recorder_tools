# stopwatch_lap_recorder_tools

A small Go package that tracks cumulative elapsed time with named laps and renders a fixed-width summary table.

## Usage

```go
package main

import (
	"fmt"
	"time"

	stopwatch "stopwatch_lap_recorder_tools"
)

func main() {
	sw := stopwatch.NewStopwatch()
	sw.Start()

	time.Sleep(100 * time.Millisecond)
	sw.Lap("setup")

	time.Sleep(50 * time.Millisecond)
	sw.Lap("work")

	fmt.Print(sw.Summary())
	fmt.Println("total:", sw.Elapsed())
}
```

The output is a plain-text table with one row per lap:

```
LAP          ELAPSED        DURATION
setup        0.100s         0.100s
work         0.150s         0.050s
```

## Why this exists

Recording laps during a run is easy; presenting them consistently is not. The common approach is to store a start time and then, at display time, recompute each lap's split by subtracting adjacent timestamps. That works until a lap is recorded out of order or a timestamp is missing, at which point the table silently shows negative or nonsensical durations.

This library makes a different trade-off: laps are recorded eagerly and each lap stores both its cumulative elapsed time and its duration since the previous lap. That means the summary is a pure function of the recorded laps and never depends on the clock at render time, which makes it trivial to test and safe to render after the stopwatch has been reset. The cost is that `Lap` must be called in chronological order for the durations to be meaningful; if you need reordering, `SortedLaps` returns a sorted copy but does not recompute durations.

## The awkward edge

Calling `Start` again while the stopwatch is already running is a no-op, not a restart. This is deliberate: a common bug is to call `Start` defensively at the top of a function and accidentally truncate elapsed time. If you want to restart, call `Reset` first. Similarly, `Lap` before `Start` does nothing, because there is no elapsed time to measure against.

## API

- `type Clock func() time.Time`
- `type Lap struct { Name string; Elapsed time.Duration; Duration time.Duration }`
- `type Stopwatch struct { ... }`
- `func NewStopwatch() *Stopwatch`
- `func NewStopwatchWithClock(clock Clock) *Stopwatch`
- `func (s *Stopwatch) Start()`
- `func (s *Stopwatch) Elapsed() time.Duration`
- `func (s *Stopwatch) Lap(name string)`
- `func (s *Stopwatch) Laps() []Lap`
- `func (s *Stopwatch) Reset()`
- `func (s *Stopwatch) Summary() string`
- `func (s *Stopwatch) SortedLaps() []Lap`

Durations in the summary are formatted with millisecond precision (`%.3fs`). Empty lap names are rendered as `lap N`, where `N` is the 1-based recording index.
