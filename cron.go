// Package gocron provides primitives to parse a Cron expression and iterate
// over the activation times.
package gocron

import "time"

type TimeUnit interface {
	// Next returns the next iteration of a schedule and `true` when valid,
	// otherwise it returns a time after `next` and `false`.
	Next(next time.Time) (time.Time, bool)
}

// Schedule is a representation of a Cron expression.
type Schedule struct {
	timeUnits []TimeUnit
}

// Parse returns a schedule from the Cron expression and returns an error if the
// syntax is not supported or incorrect.
func Parse(expression string) (Schedule, error) {
	return Parser{}.Parse(expression)
}

// Must returns a schedule from the Cron expression and panics in case of error.
func Must(expression string) Schedule {
	schedule, err := Parse(expression)
	if err != nil {
		panic(err)
	}
	return schedule
}

// Next returns a time after the given argument, but never equals to it. A zero
// time is returned when none can be found.
func (s Schedule) Next(after time.Time) (next time.Time) {
	after = after.Truncate(1 * time.Second).Add(1 * time.Second)
	return s.nextAfter(after)
}

func (s Schedule) nextAfter(after time.Time) time.Time {
	var ok bool
	for _, unit := range s.timeUnits {
		after, ok = unit.Next(after)
		if !ok {
			return s.nextAfter(after)
		}
	}
	return after
}
