package gocron

import (
	"slices"
	"time"
)

// secTimeUnit is a time unit implementation for the seconds field in a Cron
// expression.
type secTimeUnit []timeSet

// Next implements TimeUnit.
func (u secTimeUnit) Next(next time.Time) (time.Time, bool) {
	candidate, ok := searchNextCandidate(u, next, time.Time.Second, setSeconds)
	if ok {
		return candidate, true
	}

	return setMinutes(next, next.Minute()+1), false
}

func (u secTimeUnit) Previous(before time.Time) (time.Time, bool) {
	candidate, ok := searchPrevCandidate(u, before, time.Time.Second, setSeconds)
	if ok {
		return candidate, true
	}

	// Move the previous minute and set seconds to 59.
	return setSeconds(before, 0).Add(-time.Second), false
}

// minTimeUnit is a time unit implementation for the minutes field in a Cron
// expression.
type minTimeUnit []timeSet

// Next implements TimeUnit.
func (u minTimeUnit) Next(next time.Time) (time.Time, bool) {
	candidate, ok := searchNextCandidate(u, next, time.Time.Minute, setMinutes)
	if ok {
		return candidate, true
	}

	return setHours(next, next.Hour()+1), false
}

func (u minTimeUnit) Previous(before time.Time) (time.Time, bool) {
	candidate, ok := searchPrevCandidate(u, before, time.Time.Minute, setMinutes)
	if ok {
		return candidate, true
	}

	return setMinutes(before, 0).Add(-time.Second), false
}

// hourTimeUnit is a time unit implementation for the hours field in a Cron
// expression.
type hourTimeUnit []timeSet

// Next implements TimeUnit.
func (u hourTimeUnit) Next(next time.Time) (time.Time, bool) {
	candidate, ok := searchNextCandidate(u, next, time.Time.Hour, setHours)
	if ok {
		return candidate, true
	}

	return setDays(next, next.Day()+1), false
}

func (u hourTimeUnit) Previous(before time.Time) (time.Time, bool) {
	candidate, ok := searchPrevCandidate(u, before, time.Time.Hour, setHours)
	if ok {
		return candidate, true
	}

	return setHours(before, 0).Add(-time.Second), false
}

// dayTimeUnit is a time unit implementation for the days field in a Cron
// expression.
type dayTimeUnit []timeSet

// Next implements TimeUnit.
func (u dayTimeUnit) Next(next time.Time) (time.Time, bool) {
	if len(u) == 0 {
		return next, true
	}

	for _, set := range u {
		day, direction := set.NearestCandidate(next, next.Day(), true)

		switch direction {
		case hit:
			return next, true
		case inRange:
			if next = setDays(next, day); next.Day() == day {
				// Day fits inside the current month.
				return next, true
			}

			// When the day is higher than what the current month supports (e.g.
			// 30 for February).
			return setMonths(next, next.Month()), false
		}
	}

	return setMonths(next, next.Month()+1), false
}

func (u dayTimeUnit) Previous(before time.Time) (time.Time, bool) {
	candidate, ok := searchPrevCandidate(u, before, time.Time.Day, setDays)
	if ok {
		return candidate, true
	}

	return setDays(before, 1).Add(-time.Second), false
}

// monthTimeUnit is a time unit implementation for the months field in a Cron
// expression.
type monthTimeUnit []timeSet

// Next implements TimeUnit.
func (u monthTimeUnit) Next(next time.Time) (time.Time, bool) {
	candidate, ok := searchNextCandidate(u, next, time.Time.Month, setMonths)
	if ok {
		return candidate, true
	}

	return setYears(next, next.Year()+1), false
}

func (u monthTimeUnit) Previous(before time.Time) (time.Time, bool) {
	candidate, ok := searchPrevCandidate(u, before, time.Time.Month, setMonths)
	if ok {
		return candidate, true
	}

	return setMonths(before, 1).Add(-time.Second), false
}

// weekdayTimeUnit is a time unit implementation for the week days field in a
// Cron expression.
type weekdayTimeUnit []timeSet

// Next implements TimeUnit.
func (u weekdayTimeUnit) Next(next time.Time) (time.Time, bool) {
	if len(u) == 0 {
		return next, true
	}

	for _, set := range u {
		if _, direction := set.NearestCandidate(next, int(next.Weekday()), true); direction == hit {
			return next, true
		}
	}

	return setDays(next, next.Day()+1), false
}

func (u weekdayTimeUnit) Previous(before time.Time) (time.Time, bool) {
	if len(u) == 0 {
		return before, true
	}

	for i := len(u) - 1; i >= 0; i-- {
		if _, direction := u[i].NearestCandidate(before, int(before.Weekday()), false); direction == hit {
			return before, true
		}
	}

	return setDays(before, before.Day()-1), false
}

type getterFunc[T int | time.Month] func(time.Time) T
type setterFunc[T int | time.Month] func(time.Time, T) time.Time

// searchNextCandidate returns the more appropriate candidate from the time unit
// if any, otherwise it returns false.
func searchNextCandidate[T int | time.Month](
	in []timeSet, next time.Time,
	getter getterFunc[T],
	setter setterFunc[T],
) (time.Time, bool) {
	return searchCandidate[T](in, next, getter, setter, true)
}

// searchPrevCandidate returns the more appropriate candidate from the time unit
// if any, otherwise it returns false.
func searchPrevCandidate[T int | time.Month](
	in []timeSet,
	before time.Time,
	getter getterFunc[T],
	setter setterFunc[T],
) (time.Time, bool) {
	return searchCandidate(in, before, getter, setter, false)
}

func searchCandidate[T int | time.Month](
	in []timeSet,
	t time.Time,
	getter getterFunc[T],
	setter setterFunc[T],
	forwards bool,
) (time.Time, bool) {
	if len(in) == 0 {
		return t, true
	}

	var candidates []int

	for _, set := range in {
		value, direction := set.NearestCandidate(t, int(getter(t)), forwards)

		switch direction {
		case hit:
			return t, true
		case inRange:
			candidates = append(candidates, value)
		}
	}

	if len(candidates) > 0 {
		if forwards {
			// When iterating forwards, it uses the smallest candidate.
			return setter(t, T(slices.Min(candidates))), true
		}

		// When iterating backwards, it uses the biggest candidate.
		return setter(t, T(slices.Max(candidates))), true
	}

	return time.Time{}, false
}

func setSeconds(t time.Time, secs int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), secs, 0, t.Location())
}

func setMinutes(t time.Time, minutes int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), minutes, 0, 0, t.Location())
}

func setHours(t time.Time, hours int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), hours, 0, 0, 0, t.Location())
}

func setDays(t time.Time, days int) time.Time {
	return time.Date(t.Year(), t.Month(), days, 0, 0, 0, 0, t.Location())
}

func setMonths(t time.Time, months time.Month) time.Time {
	return time.Date(t.Year(), months, 1, 0, 0, 0, 0, t.Location())
}

func setYears(t time.Time, years int) time.Time {
	return time.Date(years, 1, 1, 0, 0, 0, 0, t.Location())
}
