package gocron

import "time"

// secTimeUnit is a time unit implementation for the seconds field in a Cron
// expression.
type secTimeUnit []timeSet

// Next implements TimeUnit.
func (u secTimeUnit) Next(next time.Time) (time.Time, bool) {
	if len(u) == 0 {
		return next, true
	}

	for _, set := range u {
		switch set.Compare(next, next.Second()) {
		case orderingEqual:
			return next, true
		case orderingGreater:
			return setSeconds(next, set.Value(next, next.Second())), true
		}
	}

	return setMinutes(next, next.Minute()+1), false
}

func (u secTimeUnit) Previous(before time.Time) (time.Time, bool) {
	if len(u) == 0 {
		return before, true
	}

	for i := len(u) - 1; i >= 0; i-- {
		switch u[i].Compare(before, before.Second()) {
		case orderingEqual:
			return before, true
		case orderingLess:
			return setSeconds(before, u[i].Value(before, before.Second())), true
		}
	}

	// Move the previous minute and set seconds to 59.
	return setSeconds(before, 0).Add(-time.Second), false
}

// minTimeUnit is a time unit implementation for the minutes field in a Cron
// expression.
type minTimeUnit []timeSet

// Next implements TimeUnit.
func (u minTimeUnit) Next(next time.Time) (time.Time, bool) {
	if len(u) == 0 {
		return next, true
	}

	for _, set := range u {
		switch set.Compare(next, next.Minute()) {
		case orderingEqual:
			return next, true
		case orderingGreater:
			return setMinutes(next, set.Value(next, next.Minute())), true
		}
	}

	return setHours(next, next.Hour()+1), false
}

func (u minTimeUnit) Previous(before time.Time) (time.Time, bool) {
	if len(u) == 0 {
		return before, true
	}

	for i := len(u) - 1; i >= 0; i-- {
		field := u[i]
		switch field.Compare(before, before.Minute()) {
		case orderingEqual:
			return before, true
		case orderingLess:
			return setMinutes(before, field.Value(before, before.Minute())), true
		}
	}

	return setMinutes(before, 0).Add(-time.Second), false
}

// hourTimeUnit is a time unit implementation for the hours field in a Cron
// expression.
type hourTimeUnit []timeSet

// Next implements TimeUnit.
func (u hourTimeUnit) Next(next time.Time) (time.Time, bool) {
	if len(u) == 0 {
		return next, true
	}

	for _, set := range u {
		switch set.Compare(next, next.Hour()) {
		case orderingEqual:
			return next, true
		case orderingGreater:
			return setHours(next, set.Value(next, next.Hour())), true
		}
	}

	return setDays(next, next.Day()+1), false
}

func (u hourTimeUnit) Previous(before time.Time) (time.Time, bool) {
	if len(u) == 0 {
		return before, true
	}

	for i := len(u) - 1; i >= 0; i-- {
		switch u[i].Compare(before, before.Hour()) {
		case orderingEqual:
			return before, true
		case orderingLess:
			return setHours(before, u[i].Value(before, before.Hour())), true
		}
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
		switch set.Compare(next, next.Day()) {
		case orderingEqual:
			return next, true
		case orderingGreater:
			day := set.Value(next, next.Day())

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
	if len(u) == 0 {
		return before, true
	}

	for i := len(u) - 1; i >= 0; i-- {
		switch u[i].Compare(before, before.Day()) {
		case orderingEqual:
			return before, true
		case orderingLess:
			return setDays(before, u[i].Value(before, before.Day())), true
		}
	}

	return setDays(before, 1).Add(-time.Second), false
}

// monthTimeUnit is a time unit implementation for the months field in a Cron
// expression.
type monthTimeUnit []timeSet

// Next implements TimeUnit.
func (u monthTimeUnit) Next(next time.Time) (time.Time, bool) {
	if len(u) == 0 {
		return next, true
	}

	for _, set := range u {
		switch set.Compare(next, int(next.Month())) {
		case orderingEqual:
			return next, true
		case orderingGreater:
			return setMonths(next, time.Month(set.Value(next, int(next.Month())))), true
		}
	}

	return setYears(next, next.Year()+1), false
}

func (u monthTimeUnit) Previous(before time.Time) (time.Time, bool) {
	if len(u) == 0 {
		return before, true
	}

	for i := len(u) - 1; i >= 0; i-- {
		switch u[i].Compare(before, int(before.Month())) {
		case orderingEqual:
			return before, true
		case orderingLess:
			return setMonths(before, time.Month(u[i].Value(before, int(before.Month())))), true
		}
	}

	return setMonths(before, 1).Add(-time.Second), false
}

// weekdayTimeUnit is a time unit implementation for the week days field in a
// Cron expression.
type weekdayTimeUnit sortableUnit

// Next implements TimeUnit.
func (u weekdayTimeUnit) Next(next time.Time) (time.Time, bool) {
	if len(u) == 0 {
		return next, true
	}

	for _, set := range u {
		if set.Compare(next, int(next.Weekday())) == orderingEqual {
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
		if u[i].Compare(before, int(before.Weekday())) == orderingEqual {
			return before, true
		}
	}

	return setDays(before, before.Day()-1), false
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

// sortableUnit is an implement of the sort.Interface to allow to sort the units
// in ascending order.
type sortableUnit []timeSet

func (u sortableUnit) Len() int {
	return len(u)
}

func (u sortableUnit) Less(i, j int) bool {
	return u[i].Less(u[j])
}

func (u sortableUnit) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}
