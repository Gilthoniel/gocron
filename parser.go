package gocron

import (
	"strconv"
	"strings"
	"time"
)

type Ordering int

const (
	OrderingLess Ordering = iota - 1
	OrderingEqual
	OrderingGreater
)

type ExprField interface {
	Compare(t time.Time, other int) Ordering
	Value(t time.Time) int
}

type ConvertFn func(string) (ExprField, error)

type Parser struct{}

func (p Parser) Parse(expression string) (schedule Schedule, err error) {
	matches := strings.Split(expression, " ")

	weekdays, err := p.parse(matches[5], convertWeekDay)
	if err != nil {
		return schedule, err
	}
	months, err := p.parse(matches[4], convertUnit)
	if err != nil {
		return schedule, err
	}
	days, err := p.parse(matches[3], convertWithLastDayOfMonth)
	if err != nil {
		return schedule, err
	}
	hours, err := p.parse(matches[2], convertUnit)
	if err != nil {
		return schedule, err
	}
	minutes, err := p.parse(matches[1], convertUnit)
	if err != nil {
		return schedule, err
	}
	seconds, err := p.parse(matches[0], convertUnit)
	if err != nil {
		return schedule, err
	}

	schedule.timeUnits = []TimeUnit{
		Month{units: months},
		Day{units: days},
		WeekDay{units: weekdays},
		Hour{fields: hours},
		Minute{fields: minutes},
		Second{units: seconds},
	}

	return
}

func (Parser) parse(expr string, convFn ConvertFn) (fields []ExprField, err error) {
	if expr == "*" {
		return
	}

	for _, u := range strings.Split(expr, ",") {
		var field ExprField
		switch {
		case strings.Contains(u, "-"):
			field, err = parseRange(u, convFn)
			if err != nil {
				return
			}

		default:
			field, err = convFn(u)
			if err != nil {
				return nil, err
			}
		}

		fields = append(fields, field)
	}

	return
}

func parseRange(expr string, convFn ConvertFn) (r Range, err error) {
	parts := strings.Split(expr, "-")
	r.from, err = convFn(parts[0])
	if err != nil {
		return
	}
	r.to, err = convFn(parts[1])
	if err != nil {
		return
	}

	return
}

type Second struct {
	units []ExprField
}

func (s Second) Next(next time.Time) (time.Time, bool) {
	if len(s.units) == 0 {
		// Expression is `*`.
		return next, true
	}

	for _, unit := range s.units {
		switch unit.Compare(next, next.Second()) {
		case OrderingEqual:
			return next, true
		case OrderingGreater:
			return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), unit.Value(next), 0, next.Location()), true
		}
	}

	return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute()+1, 0, 0, next.Location()), false
}

type Minute struct {
	fields []ExprField
}

func (m Minute) Next(next time.Time) (time.Time, bool) {
	if len(m.fields) == 0 {
		// Expression is `*`.
		return next, true
	}

	for _, field := range m.fields {
		switch field.Compare(next, next.Minute()) {
		case OrderingEqual:
			return next, true
		case OrderingGreater:
			return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), field.Value(next), 0, 0, next.Location()), true
		}
	}

	return time.Date(next.Year(), next.Month(), next.Day(), next.Hour()+1, 0, 0, 0, next.Location()), false
}

type Hour struct {
	fields []ExprField
}

func (h Hour) Next(next time.Time) (time.Time, bool) {
	if len(h.fields) == 0 {
		// Expression is `*`.
		return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), next.Second(), 0, next.Location()), true
	}

	for _, field := range h.fields {
		switch field.Compare(next, next.Hour()) {
		case OrderingEqual:
			return next, true
		case OrderingGreater:
			return time.Date(next.Year(), next.Month(), next.Day(), field.Value(next), 0, 0, 0, next.Location()), true
		}
	}

	return time.Date(next.Year(), next.Month(), next.Day()+1, 0, 0, 0, 0, next.Location()), false
}

type Day struct {
	units []ExprField
}

func (d Day) Next(next time.Time) (time.Time, bool) {
	if len(d.units) == 0 {
		// Expression is `*`.
		return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), next.Second(), 0, next.Location()), true
	}

	for _, unit := range d.units {
		switch unit.Compare(next, next.Day()) {
		case OrderingEqual:
			return next, true
		case OrderingGreater:
			next = time.Date(next.Year(), next.Month(), unit.Value(next), 0, 0, 0, 0, next.Location())
			if next.Day() == unit.Value(next) {
				return next, true
			} else {
				// Abort and move to the next month.
				return time.Date(next.Year(), next.Month(), 1, 0, 0, 0, 0, next.Location()), false
			}
		}
	}

	return time.Date(next.Year(), next.Month()+1, 1, 0, 0, 0, 0, next.Location()), false
}

type Month struct {
	units []ExprField
}

func (m Month) Next(next time.Time) (time.Time, bool) {
	if len(m.units) == 0 {
		// Expression is `*`.
		return next, true
	}

	for _, unit := range m.units {
		switch unit.Compare(next, int(next.Month())) {
		case OrderingEqual:
			return next, true
		case OrderingGreater:
			return time.Date(next.Year(), time.Month(unit.Value(next)), 1, 0, 0, 0, 0, next.Location()), true
		}
	}

	return time.Date(next.Year()+1, 1, 1, 0, 0, 0, 0, next.Location()), false
}

type WeekDay struct {
	units []ExprField
}

func (wd WeekDay) Next(next time.Time) (time.Time, bool) {
	if len(wd.units) == 0 {
		return next, true
	}

	for _, unit := range wd.units {
		if unit.Compare(next, int(next.Weekday())) == OrderingEqual {
			return next, true
		}
	}

	return time.Date(next.Year(), next.Month(), next.Day()+1, 0, 0, 0, 0, next.Location()), false
}

type Unit int

func (u Unit) Compare(_ time.Time, other int) Ordering {
	switch {
	case int(u) < other:
		return OrderingLess
	case int(u) == other:
		return OrderingEqual
	default:
		return OrderingGreater
	}
}

func (u Unit) Value(_ time.Time) int {
	return int(u)
}

type Range struct {
	from ExprField
	to   ExprField
}

func (r Range) Compare(t time.Time, other int) Ordering {
	switch {
	case r.to.Compare(t, other) == OrderingLess:
		return OrderingLess
	case r.from.Compare(t, other) == OrderingGreater:
		return OrderingGreater
	default:
		return OrderingEqual
	}
}

func (r Range) Value(t time.Time) int {
	return r.from.Value(t)
}

type LastDayOfMonth struct{}

func (l LastDayOfMonth) Compare(t time.Time, other int) Ordering {
	return Unit(l.Value(t)).Compare(t, other)
}

func (LastDayOfMonth) Value(t time.Time) int {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).AddDate(0, 1, -1).Day()
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

func convertWeekDay(value string) (ExprField, error) {
	if weekday, found := weekdays[strings.ToLower(value)]; found {
		return Unit(weekday), nil
	}
	return convertUnit(value)
}

func convertWithLastDayOfMonth(value string) (ExprField, error) {
	if value == "L" {
		return LastDayOfMonth{}, nil
	}
	return convertUnit(value)
}

func convertUnit(value string) (ExprField, error) {
	num, err := strconv.Atoi(value)
	return Unit(num), err
}
