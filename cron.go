// Package gocron provides primitives to parse a Cron expression and iterate
// over the activation times.
package gocron

import (
	"time"
)

var defaultParser = cronParser{}

// TimeUnit represents a single part of a Cron expression.
type TimeUnit interface {
	// Next returns the next iteration of a schedule and `true` when valid,
	// otherwise it returns a time after `next` and `false`.
	Next(next time.Time) (time.Time, bool)

	// Previous returns the previous iteration of a schedule and `true` when
	// valid, otherwise it returns a time before `prev` and `false`.
	Previous(prev time.Time) (time.Time, bool)
}

// Schedule is a representation of a Cron expression.
type Schedule struct {
	timeUnits []TimeUnit
}

// Parse returns a schedule from the Cron expression and returns an error if the
// syntax is not supported or incorrect.
func Parse(expression string) (Schedule, error) {
	return defaultParser.Parse(expression)
}

// MustParse returns a schedule from the Cron expression and panics in case of
// error.
func MustParse(expression string) Schedule {
	schedule, err := Parse(expression)
	if err != nil {
		panic(err)
	}
	return schedule
}

// Next returns a time after the given argument, but never equals to it. A zero
// time is returned when none can be found.
func (s Schedule) Next(after time.Time) (next time.Time) {
	next = after.Truncate(1 * time.Second).Add(1 * time.Second)
	var ok bool

	for !ok {
		if next.IsZero() || next.Year() > rangeMaxYear {
			// Return a zero time when the expression is unable to find a proper
			// time after a given number of years.
			return time.Time{}
		}
		next, ok = s.nextAfter(next)
	}

	return
}

func (s Schedule) nextAfter(after time.Time) (time.Time, bool) {
	var ok bool
	for _, unit := range s.timeUnits {
		after, ok = unit.Next(after)
		if !ok {
			return after, false
		}
	}
	return after, true
}

// Previous returns a time before the given argument, but never equals it. A
// zero time is returned when none can be found.
func (s Schedule) Previous(before time.Time) (prev time.Time) {
	prev = before.Truncate(1 * time.Second)
	if prev.Equal(before) {
		prev = prev.Add(-time.Second)
	}

	var ok bool

	for !ok {
		if prev.IsZero() || prev.Year() > rangeMaxYear {
			return time.Time{}
		}
		prev, ok = s.prevBefore(prev)
	}

	return
}

func (s Schedule) prevBefore(before time.Time) (time.Time, bool) {
	var ok bool
	for _, unit := range s.timeUnits {
		before, ok = unit.Previous(before)
		if !ok {
			return before, false
		}
	}
	return before, true
}

// Upcoming returns an iterator that will iterate over the activation times of
// the Cron expression of the schedule forwards from the given time.
func (s Schedule) Upcoming(after time.Time) *Iterator {
	i := &Iterator{fn: s.Next, next: after}
	i.Next() // initialize the first value of the iterator.
	return i
}

// Preceding returns an iterator that will iterate over the actiation times of
// the Cron expression of the schedule backwards from the given time.
func (s Schedule) Preceding(before time.Time) *Iterator {
	i := &Iterator{fn: s.Previous, next: before}
	i.Next() // initialize the first value of the iterator.
	return i
}

type Iterator struct {
	fn   func(time.Time) time.Time
	next time.Time
}

// HasNext returns true if an activation time is available. When it returns
// false, any call to `Next` will return a zero time.
func (i *Iterator) HasNext() bool {
	return !i.next.IsZero()
}

// Next returns the next activation time of the schedule, or a zero time if none
// is available.
func (i *Iterator) Next() (next time.Time) {
	next = i.next
	i.next = i.fn(next)
	return
}
