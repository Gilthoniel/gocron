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
	Value(t time.Time, other int) int
}

type ConvertFn func(string) (ExprField, error)

// Parser is a parser from Cron expressions.
type Parser struct{}

func (p Parser) Parse(expression string) (schedule Schedule, err error) {
	matches := strings.Split(expression, " ")

	weekdays, err := p.parse(matches[5], convertWeekDay, 0, 6)
	if err != nil {
		return schedule, err
	}
	months, err := p.parse(matches[4], convertUnit, 1, 12)
	if err != nil {
		return schedule, err
	}
	days, err := p.parse(matches[3], convertWithLastDayOfMonth, 1, 31)
	if err != nil {
		return schedule, err
	}
	hours, err := p.parse(matches[2], convertUnit, 0, 23)
	if err != nil {
		return schedule, err
	}
	minutes, err := p.parse(matches[1], convertUnit, 0, 59)
	if err != nil {
		return schedule, err
	}
	seconds, err := p.parse(matches[0], convertUnit, 0, 59)
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

func (Parser) parse(expr string, convFn ConvertFn, min, max int) (fields []ExprField, err error) {
	if expr == "*" {
		return
	}

	for _, u := range strings.Split(expr, ",") {
		var field ExprField

		switch {
		case strings.Contains(u, "/"):
			field, err = parseInterval(u, convFn, min, max)
		case strings.Contains(u, "-"):
			field, err = parseRange(u, convFn)
		default:
			field, err = convFn(u)
		}

		if err != nil {
			return
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

func parseInterval(expr string, convFn ConvertFn, min, max int) (i Interval, err error) {
	parts := strings.Split(expr, "/")

	i.incr, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}

	switch {
	case parts[0] == "*":
		i.rge.from = Unit(min)
		i.rge.to = Unit(max)
	case strings.Contains(parts[0], "-"):
		i.rge, err = parseRange(parts[0], convFn)
		if err != nil {
			return
		}
	default:
		i.rge.from, err = convFn(parts[0])
		if err != nil {
			return
		}

		i.rge.to = Unit(max)
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
			return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), unit.Value(next, next.Second()), 0, next.Location()), true
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
			return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), field.Value(next, next.Minute()), 0, 0, next.Location()), true
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
			return time.Date(next.Year(), next.Month(), next.Day(), field.Value(next, next.Hour()), 0, 0, 0, next.Location()), true
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
			return time.Date(next.Year(), time.Month(unit.Value(next, int(next.Month()))), 1, 0, 0, 0, 0, next.Location()), true
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

// Unit is an expression field that represents a single possible value.
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

func (u Unit) Value(_ time.Time, _ int) int {
	return int(u)
}

// Range is an expression field that represents an inclusive range of values.
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

func (r Range) Value(t time.Time, other int) int {
	return r.from.Value(t, other)
}

type Interval struct {
	rge  Range
	incr int
}

func (i Interval) Compare(t time.Time, other int) Ordering {
	nearestAfter := i.Value(t, other)
	isMaxGreaterOrEqual := i.rge.to.Compare(t, nearestAfter) != OrderingLess

	switch {
	case nearestAfter > other && isMaxGreaterOrEqual:
		// nearestAfter is after `other` and in the range.
		return OrderingGreater
	case other == nearestAfter && isMaxGreaterOrEqual:
		// nearestAfter is equal to `other` and in the range.
		return OrderingEqual
	default:
		return OrderingLess
	}
}

func (i Interval) Value(t time.Time, other int) int {
	from := i.rge.from.Value(t, other)
	if other < from {
		return from
	}
	remainder := (other - from) % i.incr
	if remainder > 0 {
		other += i.incr - remainder
	}

	return other
}

// LastDayOfMonth is a specialized expression field to determine the last day of
// a month.
type LastDayOfMonth struct{}

func (l LastDayOfMonth) Compare(t time.Time, other int) Ordering {
	return Unit(l.Value(t, other)).Compare(t, other)
}

func (LastDayOfMonth) Value(t time.Time, _ int) int {
	return findLastDayOfMonth(t).Day()
}

// LastWeekDayOfMonth is a specialized expression field to determine which day
// of the month corresponds to the last occurence of a week day.
type LastWeekDayOfMonth struct {
	weekday time.Weekday
}

func (l LastWeekDayOfMonth) Compare(t time.Time, other int) Ordering {
	unit := Unit(l.Value(t, other))
	return unit.Compare(t, t.Day())
}

func (l LastWeekDayOfMonth) Value(t time.Time, _ int) int {
	lastDayOfMonth := findLastDayOfMonth(t)
	diff := int(lastDayOfMonth.Weekday() - l.weekday)
	if diff < 0 {
		diff += 7
	}
	return lastDayOfMonth.Day() - diff
}

// findLastDayOfMonth returns a time corresponding to the last of the month at
// midnight.
func findLastDayOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).AddDate(0, 1, -1)
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
	if value == "L" {
		return Unit(time.Saturday), nil
	}
	if strings.HasSuffix(value, "L") {
		weekday, err := strconv.Atoi(strings.TrimSuffix(value, "L"))
		if err != nil {
			return nil, err
		}

		return LastWeekDayOfMonth{weekday: time.Weekday(weekday)}, nil
	}
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
