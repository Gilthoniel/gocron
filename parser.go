package gocron

import (
	"strconv"
	"strings"
	"time"
)

type Parser struct{}

func (p Parser) Parse(expression string) (schedule Schedule, err error) {
	matches := strings.Split(expression, " ")

	weekdays, err := p.parse(matches[5], convertWeekDay)
	if err != nil {
		return schedule, err
	}
	months, err := p.parse(matches[4], strconv.Atoi)
	if err != nil {
		return schedule, err
	}
	days, err := p.parse(matches[3], strconv.Atoi)
	if err != nil {
		return schedule, err
	}
	hours, err := p.parse(matches[2], strconv.Atoi)
	if err != nil {
		return schedule, err
	}
	minutes, err := p.parse(matches[1], strconv.Atoi)
	if err != nil {
		return schedule, err
	}
	seconds, err := p.parse(matches[0], strconv.Atoi)
	if err != nil {
		return schedule, err
	}

	schedule.timeUnits = []TimeUnit{
		Month{units: months},
		Day{units: days},
		WeekDay{units: weekdays},
		Hour{units: hours},
		Minute{units: minutes},
		Second{units: seconds},
	}

	return
}

func (Parser) parse(expr string, convFn func(string) (int, error)) (units []Unit, err error) {
	if expr == "*" {
		return
	}

	var value int

	for _, u := range strings.Split(expr, ",") {
		if value, err = convFn(u); err != nil {
			return nil, err
		}

		units = append(units, Unit(value))
	}

	return
}

type Second struct {
	units []Unit
}

func (s Second) Next(next time.Time) (time.Time, bool) {
	if len(s.units) == 0 {
		// Expression is `*`.
		return next, true
	}

	for _, unit := range s.units {
		if unit.Equal(next.Second()) {
			return next, true
		}
		if unit.GreaterThan(next.Second()) {
			return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), int(unit), 0, next.Location()), true
		}
	}

	return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute()+1, 0, 0, next.Location()), false
}

type Minute struct {
	units []Unit
}

func (m Minute) Next(next time.Time) (time.Time, bool) {
	if len(m.units) == 0 {
		// Expression is `*`.
		return next, true
	}

	for _, unit := range m.units {
		if unit.Equal(next.Minute()) {
			return next, true
		}
		if unit.GreaterThan(next.Minute()) {
			return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), int(unit), 0, 0, next.Location()), true
		}
	}

	return time.Date(next.Year(), next.Month(), next.Day(), next.Hour()+1, 0, 0, 0, next.Location()), false
}

type Hour struct {
	units []Unit
}

func (h Hour) Next(next time.Time) (time.Time, bool) {
	if len(h.units) == 0 {
		// Expression is `*`.
		return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), next.Second(), 0, next.Location()), true
	}

	for _, unit := range h.units {
		if unit.Equal(next.Hour()) {
			return next, true
		}
		if unit.GreaterThan(next.Hour()) {
			return time.Date(next.Year(), next.Month(), next.Day(), int(unit), 0, 0, 0, next.Location()), true
		}
	}

	return time.Date(next.Year(), next.Month(), next.Day()+1, 0, 0, 0, 0, next.Location()), false
}

type Day struct {
	units []Unit
}

func (d Day) Next(next time.Time) (time.Time, bool) {
	if len(d.units) == 0 {
		// Expression is `*`.
		return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), next.Second(), 0, next.Location()), true
	}

	for _, unit := range d.units {
		if unit.Equal(next.Day()) {
			return next, true
		}
		if unit.GreaterThan(next.Day()) {
			next = time.Date(next.Year(), next.Month(), int(unit), 0, 0, 0, 0, next.Location())
			if next.Day() == int(unit) {
				return next, true
			} else {
				// Abort and move to the next month.
				return time.Date(next.Year(), next.Month()+1, 1, 0, 0, 0, 0, next.Location()), false
			}
		}
	}

	return time.Date(next.Year(), next.Month()+1, 1, 0, 0, 0, 0, next.Location()), false
}

type Month struct {
	units []Unit
}

func (m Month) Next(next time.Time) (time.Time, bool) {
	if len(m.units) == 0 {
		// Expression is `*`.
		return next, true
	}

	for _, unit := range m.units {
		if unit.Equal(int(next.Month())) {
			return next, true
		}
		if unit.GreaterThan(int(next.Month())) {
			return time.Date(next.Year(), time.Month(unit), 0, 0, 0, 0, 0, next.Location()), true
		}
	}

	return time.Date(next.Year()+1, 1, 1, 0, 0, 0, 0, next.Location()), false
}

type WeekDay struct {
	units []Unit
}

func (wd WeekDay) Next(next time.Time) (time.Time, bool) {
	if len(wd.units) == 0 {
		return next, true
	}

	for _, unit := range wd.units {
		if unit.Equal(int(next.Weekday())) {
			return next, true
		}
	}

	return time.Date(next.Year(), next.Month(), next.Day()+1, 0, 0, 0, 0, next.Location()), false
}

type Unit int

func (u Unit) Equal(other int) bool {
	return int(u) == other
}

func (u Unit) GreaterThan(other int) bool {
	return int(u) > other
}

var weekdays = map[string]time.Weekday{
	"sun": time.Sunday,
	"mon": time.Monday,
	"tue": time.Tuesday,
	"wed": time.Wednesday,
	"thu": time.Thursday,
	"fri": time.Friday,
	"sat": time.Saturday,
}

func convertWeekDay(value string) (int, error) {
	if weekday, found := weekdays[strings.ToLower(value)]; found {
		return int(weekday), nil
	}
	return strconv.Atoi(value)
}
