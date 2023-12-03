package gocron

import (
	"strconv"
	"strings"
	"time"
)

// Parser is a parser from Cron expressions.
type Parser struct{}

func (p Parser) Parse(expression string) (schedule Schedule, err error) {
	matches := strings.Split(expression, " ")

	weekdays, err := p.parse(matches[5], convertWeekDay, 0, 6)
	if err != nil {
		return schedule, newTimeUnitErr(WeekDays, err)
	}
	months, err := p.parse(matches[4], convertUnit, 1, 12)
	if err != nil {
		return schedule, newTimeUnitErr(Months, err)
	}
	days, err := p.parse(matches[3], convertWithLastDayOfMonth, 1, 31)
	if err != nil {
		return schedule, newTimeUnitErr(Days, err)
	}
	hours, err := p.parse(matches[2], convertUnit, 0, 23)
	if err != nil {
		return schedule, newTimeUnitErr(Hours, err)
	}
	minutes, err := p.parse(matches[1], convertUnit, 0, 59)
	if err != nil {
		return schedule, newTimeUnitErr(Minutes, err)
	}
	seconds, err := p.parse(matches[0], convertUnit, 0, 59)
	if err != nil {
		return schedule, newTimeUnitErr(Seconds, err)
	}

	if isNotSpecified(weekdays) && isNotSpecified(days) {
		err = ErrMultipleNotSpecified
	}

	schedule.timeUnits = []TimeUnit{
		monthTimeUnit{units: months},
		dayTimeUnit{units: days},
		weekdayTimeUnit{units: weekdays},
		hourTimeUnit{fields: hours},
		minTimeUnit{fields: minutes},
		secTimeUnit{units: seconds},
	}

	return
}

func (Parser) parse(expr string, convFn converterFn, min, max int) (fields []exprField, err error) {
	if expr == "*" {
		return
	}
	if expr == "?" {
		fields = append(fields, notSpecifiedExpr{})
		return
	}

	for _, u := range strings.Split(expr, ",") {
		var field exprField

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
		if !field.SubsetOf(min, max) {
			err = ErrValueOutsideRange
			return
		}

		fields = append(fields, field)
	}

	return
}

func parseRange(expr string, convFn converterFn) (r rangeExpr, err error) {
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

func parseInterval(expr string, convFn converterFn, min, max int) (i intervalExpr, err error) {
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

// isNotSpecified returns true if any of the field is not specified.
func isNotSpecified(fields []exprField) bool {
	for _, field := range fields {
		if _, ok := field.(notSpecifiedExpr); ok {
			return true
		}
	}
	return false
}

// notSpecifiedExpr is an expression field that represent a question mark `?`,
// or in other words a value not specified.
type notSpecifiedExpr struct{}

func (notSpecifiedExpr) Compare(_ time.Time, other int) ordering {
	return orderingEqual
}

func (notSpecifiedExpr) Value(_ time.Time, other int) int {
	return other
}

func (notSpecifiedExpr) SubsetOf(_, _ int) bool {
	// A field not specified is always contained by any range of values.
	return true
}

// unitExpr is an expression field that represents a single possible value.
type unitExpr int

func (u unitExpr) Compare(_ time.Time, other int) ordering {
	switch {
	case int(u) < other:
		return orderingLess
	case int(u) == other:
		return orderingEqual
	default:
		return orderingGreater
	}
}

func (u unitExpr) Value(_ time.Time, _ int) int {
	return int(u)
}

func (u unitExpr) SubsetOf(min, max int) bool {
	return int(u) >= min && int(u) <= max
}

// rangeExpr is an expression field that represents an inclusive range of
// values.
type rangeExpr struct {
	from exprField
	to   exprField
}

func (r rangeExpr) Compare(t time.Time, other int) ordering {
	switch {
	case r.to.Compare(t, other) == orderingLess:
		return orderingLess
	case r.from.Compare(t, other) == orderingGreater:
		return orderingGreater
	default:
		return orderingEqual
	}
}

func (r rangeExpr) Value(t time.Time, other int) int {
	return r.from.Value(t, other)
}

func (r rangeExpr) SubsetOf(min, max int) bool {
	return r.from.SubsetOf(min, max) && r.to.SubsetOf(min, max)
}

type intervalExpr struct {
	rge  rangeExpr
	incr int
}

func (i intervalExpr) Compare(t time.Time, other int) ordering {
	nearestAfter := i.Value(t, other)
	isMaxGreaterOrEqual := i.rge.to.Compare(t, nearestAfter) != orderingLess

	switch {
	case nearestAfter > other && isMaxGreaterOrEqual:
		// nearestAfter is after `other` and in the range.
		return orderingGreater
	case other == nearestAfter && isMaxGreaterOrEqual:
		// nearestAfter is equal to `other` and in the range.
		return orderingEqual
	default:
		return orderingLess
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

func (i intervalExpr) SubsetOf(min, max int) bool {
	return i.rge.SubsetOf(min, max)
}

// lastDayOfMonthExpr is a specialized expression field to determine the last
// day of a month.
type lastDayOfMonthExpr struct{}

func (l lastDayOfMonthExpr) Compare(t time.Time, other int) ordering {
	return unitExpr(l.Value(t, other)).Compare(t, other)
}

func (lastDayOfMonthExpr) Value(t time.Time, _ int) int {
	return findLastDayOfMonth(t).Day()
}

func (lastDayOfMonthExpr) SubsetOf(_, _ int) bool {
	// the value is always calculated to lie inside the proper range.
	return true
}

// lastWeekDayOfMonthExpr is a specialized expression field to determine which
// day of the month corresponds to the last occurence of a week day.
type lastWeekDayOfMonthExpr struct {
	weekday time.Weekday
}

func (l lastWeekDayOfMonthExpr) Compare(t time.Time, other int) ordering {
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

func (lastWeekDayOfMonthExpr) SubsetOf(_, _ int) bool {
	// the value is always calculated to lie inside the proper range.
	return true
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

func convertWeekDay(value string) (exprField, error) {
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

func convertWithLastDayOfMonth(value string) (exprField, error) {
	if value == "L" {
		return lastDayOfMonthExpr{}, nil
	}
	return convertUnit(value)
}

func convertUnit(value string) (exprField, error) {
	num, err := strconv.Atoi(value)
	return unitExpr(num), err
}

type ordering int

const (
	orderingLess ordering = iota - 1
	orderingEqual
	orderingGreater
)

type exprField interface {
	Compare(t time.Time, other int) ordering
	Value(t time.Time, other int) int
	SubsetOf(min, max int) bool
}

type converterFn func(string) (exprField, error)
