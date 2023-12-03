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

	schedule.timeUnits = []timeUnit{
		monthTimeUnit{units: months},
		dayTimeUnit{units: days},
		weekdayTimeUnit{units: weekdays},
		hourTimeUnit{fields: hours},
		minTimeUnit{fields: minutes},
		secTimeUnit{units: seconds},
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

func parseRange(expr string, convFn ConvertFn) (r rangeExpr, err error) {
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

func parseInterval(expr string, convFn ConvertFn, min, max int) (i intervalExpr, err error) {
	parts := strings.Split(expr, "/")

	i.incr, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}

	switch {
	case parts[0] == "*":
		i.rge.from = unitExpr(min)
		i.rge.to = unitExpr(max)
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

		i.rge.to = unitExpr(max)
	}

	return
}

// unitExpr is an expression field that represents a single possible value.
type unitExpr int

func (u unitExpr) Compare(_ time.Time, other int) Ordering {
	switch {
	case int(u) < other:
		return OrderingLess
	case int(u) == other:
		return OrderingEqual
	default:
		return OrderingGreater
	}
}

func (u unitExpr) Value(_ time.Time, _ int) int {
	return int(u)
}

// rangeExpr is an expression field that represents an inclusive range of
// values.
type rangeExpr struct {
	from ExprField
	to   ExprField
}

func (r rangeExpr) Compare(t time.Time, other int) Ordering {
	switch {
	case r.to.Compare(t, other) == OrderingLess:
		return OrderingLess
	case r.from.Compare(t, other) == OrderingGreater:
		return OrderingGreater
	default:
		return OrderingEqual
	}
}

func (r rangeExpr) Value(t time.Time, other int) int {
	return r.from.Value(t, other)
}

type intervalExpr struct {
	rge  rangeExpr
	incr int
}

func (i intervalExpr) Compare(t time.Time, other int) Ordering {
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

func (i intervalExpr) Value(t time.Time, other int) int {
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

// lastDayOfMonthExpr is a specialized expression field to determine the last
// day of a month.
type lastDayOfMonthExpr struct{}

func (l lastDayOfMonthExpr) Compare(t time.Time, other int) Ordering {
	return unitExpr(l.Value(t, other)).Compare(t, other)
}

func (lastDayOfMonthExpr) Value(t time.Time, _ int) int {
	return findLastDayOfMonth(t).Day()
}

// lastWeekDayOfMonthExpr is a specialized expression field to determine which
// day of the month corresponds to the last occurence of a week day.
type lastWeekDayOfMonthExpr struct {
	weekday time.Weekday
}

func (l lastWeekDayOfMonthExpr) Compare(t time.Time, other int) Ordering {
	unit := unitExpr(l.Value(t, other))
	return unit.Compare(t, t.Day())
}

func (l lastWeekDayOfMonthExpr) Value(t time.Time, _ int) int {
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
		return unitExpr(time.Saturday), nil
	}
	if strings.HasSuffix(value, "L") {
		weekday, err := strconv.Atoi(strings.TrimSuffix(value, "L"))
		if err != nil {
			return nil, err
		}

		return lastWeekDayOfMonthExpr{weekday: time.Weekday(weekday)}, nil
	}
	if weekday, found := weekdays[strings.ToLower(value)]; found {
		return unitExpr(weekday), nil
	}
	return convertUnit(value)
}

func convertWithLastDayOfMonth(value string) (ExprField, error) {
	if value == "L" {
		return lastDayOfMonthExpr{}, nil
	}
	return convertUnit(value)
}

func convertUnit(value string) (ExprField, error) {
	num, err := strconv.Atoi(value)
	return unitExpr(num), err
}
