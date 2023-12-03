package gocron

import "time"

// secTimeUnit is a time unit implementation for the seconds field in a Cron
// expression.
type secTimeUnit struct {
	units []exprField
}

// Next implements TimeUnit.
func (s secTimeUnit) Next(next time.Time) (time.Time, bool) {
	if len(s.units) == 0 {
		// Expression is `*`.
		return next, true
	}

	for _, unit := range s.units {
		switch unit.Compare(next, next.Second()) {
		case orderingEqual:
			return next, true
		case orderingGreater:
			return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), unit.Value(next, next.Second()), 0, next.Location()), true
		}
	}

	return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute()+1, 0, 0, next.Location()), false
}

// minTimeUnit is a time unit implementation for the minutes field in a Cron
// expression.
type minTimeUnit struct {
	fields []exprField
}

// Next implements TimeUnit.
func (m minTimeUnit) Next(next time.Time) (time.Time, bool) {
	if len(m.fields) == 0 {
		// Expression is `*`.
		return next, true
	}

	for _, field := range m.fields {
		switch field.Compare(next, next.Minute()) {
		case orderingEqual:
			return next, true
		case orderingGreater:
			return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), field.Value(next, next.Minute()), 0, 0, next.Location()), true
		}
	}

	return time.Date(next.Year(), next.Month(), next.Day(), next.Hour()+1, 0, 0, 0, next.Location()), false
}

// hourTimeUnit is a time unit implementation for the hours field in a Cron
// expression.
type hourTimeUnit struct {
	fields []exprField
}

// Next implements TimeUnit.
func (h hourTimeUnit) Next(next time.Time) (time.Time, bool) {
	if len(h.fields) == 0 {
		// Expression is `*`.
		return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), next.Second(), 0, next.Location()), true
	}

	for _, field := range h.fields {
		switch field.Compare(next, next.Hour()) {
		case orderingEqual:
			return next, true
		case orderingGreater:
			return time.Date(next.Year(), next.Month(), next.Day(), field.Value(next, next.Hour()), 0, 0, 0, next.Location()), true
		}
	}

	return time.Date(next.Year(), next.Month(), next.Day()+1, 0, 0, 0, 0, next.Location()), false
}

// dayTimeUnit is a time unit implementation for the days field in a Cron
// expression.
type dayTimeUnit struct {
	units []exprField
}

// Next implements TimeUnit.
func (d dayTimeUnit) Next(next time.Time) (time.Time, bool) {
	if len(d.units) == 0 {
		// Expression is `*`.
		return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), next.Second(), 0, next.Location()), true
	}

	for _, unit := range d.units {
		switch unit.Compare(next, next.Day()) {
		case orderingEqual:
			return next, true
		case orderingGreater:
			next = time.Date(next.Year(), next.Month(), unit.Value(next, next.Day()), 0, 0, 0, 0, next.Location())
			if next.Day() == unit.Value(next, next.Day()) {
				return next, true
			} else {
				// Abort and move to the next month.
				return time.Date(next.Year(), next.Month(), 1, 0, 0, 0, 0, next.Location()), false
			}
		}
	}

	return time.Date(next.Year(), next.Month()+1, 1, 0, 0, 0, 0, next.Location()), false
}

// monthTimeUnit is a time unit implementation for the months field in a Cron
// expression.
type monthTimeUnit struct {
	units []exprField
}

// Next implements TimeUnit.
func (m monthTimeUnit) Next(next time.Time) (time.Time, bool) {
	if len(m.units) == 0 {
		// Expression is `*`.
		return next, true
	}

	for _, unit := range m.units {
		switch unit.Compare(next, int(next.Month())) {
		case orderingEqual:
			return next, true
		case orderingGreater:
			return time.Date(next.Year(), time.Month(unit.Value(next, int(next.Month()))), 1, 0, 0, 0, 0, next.Location()), true
		}
	}

	return time.Date(next.Year()+1, 1, 1, 0, 0, 0, 0, next.Location()), false
}

// weekdayTimeUnit is a time unit implementation for the week days field in a
// Cron expression.
type weekdayTimeUnit struct {
	units []exprField
}

// Next implements TimeUnit.
func (wd weekdayTimeUnit) Next(next time.Time) (time.Time, bool) {
	if len(wd.units) == 0 {
		return next, true
	}

	for _, unit := range wd.units {
		if unit.Compare(next, int(next.Weekday())) == orderingEqual {
			return next, true
		}
	}

	return time.Date(next.Year(), next.Month(), next.Day()+1, 0, 0, 0, 0, next.Location()), false
}
