package gocron

import (
	"strconv"
	"strings"
	"time"
)

// cronParser is a parser from Cron expressions.
type cronParser struct{}

func (p cronParser) Parse(expression string) (schedule Schedule, err error) {
	matches := strings.Split(expression, " ")
	if len(matches) != 6 {
		return schedule, ErrMalformedExpression
	}

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
		monthTimeUnit(months),
		dayTimeUnit(days),
		weekdayTimeUnit(weekdays),
		hourTimeUnit(hours),
		minTimeUnit(minutes),
		secTimeUnit(seconds),
	}

	return
}

func (cronParser) parse(expr string, convFn converterFn, min, max int) (fields []timeSet, err error) {
	if expr == "*" {
		return
	}
	if expr == "?" {
		fields = append(fields, notSpecifiedExpr{})
		return
	}

	for _, u := range strings.Split(expr, ",") {
		var field timeSet

		switch {
		case strings.Contains(u, "/"):
			field, err = parseInterval(u, convFn, min, max)
		case strings.Contains(u, "-") && !strings.Contains(u, "L-"):
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
	if len(parts) != 2 {
		err = ErrMalformedField
		return
	}

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
	if len(parts) != 2 {
		err = ErrMalformedField
		return
	}

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
func isNotSpecified(fields []timeSet) bool {
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

func (notSpecifiedExpr) Nearest(t time.Time, other int) (int, ordering) {
	return other, orderingEqual
}

func (notSpecifiedExpr) SubsetOf(_, _ int) bool {
	// A field not specified is always contained by any range of values.
	return true
}

// unitExpr is an expression field that represents a single possible value.
type unitExpr int

func (u unitExpr) Nearest(t time.Time, other int) (int, ordering) {
	value := int(u)

	switch {
	case value < other:
		return value, orderingLess
	case value == other:
		return value, orderingEqual
	default:
		return value, orderingGreater
	}
}

func (u unitExpr) SubsetOf(min, max int) bool {
	return int(u) >= min && int(u) <= max
}

// rangeExpr is an expression field that represents an inclusive range of
// values.
type rangeExpr struct {
	from timeSet
	to   timeSet
}

func (r rangeExpr) Nearest(t time.Time, other int) (int, ordering) {
	value, _ := r.from.Nearest(t, other)

	if _, direction := r.to.Nearest(t, other); direction == orderingLess {
		return value, orderingLess
	}
	if _, direction := r.from.Nearest(t, other); direction == orderingGreater {
		return value, orderingGreater
	}

	return value, orderingEqual
}

func (r rangeExpr) SubsetOf(min, max int) bool {
	return r.from.SubsetOf(min, max) && r.to.SubsetOf(min, max)
}

type intervalExpr struct {
	rge  rangeExpr
	incr int
}

func (i intervalExpr) Nearest(t time.Time, other int) (int, ordering) {
	nearestAfter := i.valueFor(t, other)

	_, direction := i.rge.to.Nearest(t, nearestAfter)
	isMaxGreaterOrEqual := direction != orderingLess

	switch {
	case nearestAfter > other && isMaxGreaterOrEqual:
		// nearestAfter is after `other` and in the range.
		return nearestAfter, orderingGreater
	case other == nearestAfter && isMaxGreaterOrEqual:
		// nearestAfter is equal to `other` and in the range.
		return nearestAfter, orderingEqual
	default:
		return nearestAfter, orderingLess
	}
}

func (i intervalExpr) valueFor(t time.Time, other int) int {
	from, _ := i.rge.from.Nearest(t, other)
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

// nthLastDayOfMonthExpr is a specialized expression field to determine the last
// day of a month.
type nthLastDayOfMonthExpr struct {
	nthLast int
}

func (e nthLastDayOfMonthExpr) Nearest(t time.Time, other int) (int, ordering) {
	value := findLastDayOfMonth(t).Day() - e.nthLast
	_, direction := unitExpr(value).Nearest(t, other)
	return value, direction
}

func (e nthLastDayOfMonthExpr) SubsetOf(min, max int) bool {
	return e.nthLast+1 >= min && max-e.nthLast >= min
}

// lastWeekDayOfMonthExpr is a specialized expression field to determine which
// day of the month corresponds to the last occurence of a week day.
type lastWeekDayOfMonthExpr struct {
	weekday time.Weekday
}

func (l lastWeekDayOfMonthExpr) Nearest(t time.Time, _ int) (int, ordering) {
	value := l.valueFor(t)
	_, direction := unitExpr(value).Nearest(t, t.Day())
	return value, direction
}

func (l lastWeekDayOfMonthExpr) valueFor(t time.Time) int {
	lastDayOfMonth := findLastDayOfMonth(t)
	diff := int(lastDayOfMonth.Weekday() - l.weekday)
	if diff < 0 {
		// Add a week length to get a positive value.
		diff += int(time.Saturday) + 1
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

func convertWeekDay(value string) (timeSet, error) {
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

func convertWithLastDayOfMonth(value string) (timeSet, error) {
	if value == "L" {
		return nthLastDayOfMonthExpr{}, nil
	}
	if strings.HasPrefix(value, "L-") {
		nth, err := strconv.Atoi(strings.TrimPrefix(value, "L-"))
		return nthLastDayOfMonthExpr{nthLast: nth - 1}, err
	}
	return convertUnit(value)
}

func convertUnit(value string) (timeSet, error) {
	num, err := strconv.Atoi(value)
	return unitExpr(num), err
}

type ordering int

const (
	orderingLess ordering = iota - 1
	orderingEqual
	orderingGreater
)

// timeSet represents a set of possible values for a time unit.
type timeSet interface {
	// Nearest returns the value and its direction compared the given time and
	// value.
	Nearest(time.Time, int) (int, ordering)

	// SubsetOf returns true if the time set is included in the range [min,
	// max].
	SubsetOf(min, max int) bool
}

type converterFn func(string) (timeSet, error)
