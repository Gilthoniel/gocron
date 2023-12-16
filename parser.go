package gocron

import (
	"strconv"
	"strings"
	"time"
)

const (
	minExprMatches    = 6
	maxExprMatches    = 7
	rangeSplitSize    = 2
	intervalSplitSize = 2
	nthSplitSize      = 2

	rangeMinYear       = 1
	rangeMaxYear       = 9999
	rangeMinWeekday    = 0
	rangeMaxWeekday    = 6
	rangeMinMonth      = 1
	rangeMaxMonth      = 12
	rangeMinDayOfMonth = 1
	rangeMaxDayOfMonth = 31
	rangeMinHour       = 0
	rangeMaxHour       = 23
	rangeMinMinute     = 0
	rangeMaxMinute     = 59
	rangeMinSecond     = 0
	rangeMaxSecond     = 59
)

// cronParser is a parser from Cron expressions.
type cronParser struct{}

func (p cronParser) Parse(expression string) (schedule Schedule, err error) {
	matches := strings.Split(expression, " ")
	if len(matches) < minExprMatches || len(matches) > maxExprMatches {
		return schedule, ErrMalformedExpression
	}

	weekdays, err := p.parse(matches[5], convertWeekDay, rangeMinWeekday, rangeMaxWeekday)
	if err != nil {
		return schedule, newTimeUnitErr(WeekDays, err)
	}
	months, err := p.parse(matches[4], convertUnit, rangeMinMonth, rangeMaxMonth)
	if err != nil {
		return schedule, newTimeUnitErr(Months, err)
	}
	days, err := p.parse(matches[3], convertWithLastDayOfMonth, rangeMinDayOfMonth, rangeMaxDayOfMonth)
	if err != nil {
		return schedule, newTimeUnitErr(Days, err)
	}
	hours, err := p.parse(matches[2], convertUnit, rangeMinHour, rangeMaxHour)
	if err != nil {
		return schedule, newTimeUnitErr(Hours, err)
	}
	minutes, err := p.parse(matches[1], convertUnit, rangeMinMinute, rangeMaxMinute)
	if err != nil {
		return schedule, newTimeUnitErr(Minutes, err)
	}
	seconds, err := p.parse(matches[0], convertUnit, rangeMinSecond, rangeMaxSecond)
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

	if len(matches) == maxExprMatches {
		years, err := p.parse(matches[maxExprMatches-1], convertUnit, rangeMinYear, rangeMaxYear)
		if err != nil {
			return schedule, newTimeUnitErr(Years, err)
		}

		schedule.timeUnits = append([]TimeUnit{yearTimeUnit(years)}, schedule.timeUnits...)
	}

	return schedule, err
}

func (cronParser) parse(expr string, convFn converterFn, min, max int) (fields []timeSet, err error) {
	if expr == "*" {
		return nil, err
	}
	if expr == "?" {
		fields = append(fields, notSpecifiedExpr{})
		return fields, err
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
			return nil, err
		}
		if !field.SubsetOf(min, max) {
			return nil, ErrValueOutsideRange
		}

		fields = append(fields, field)
	}

	return fields, nil
}

func parseRange(expr string, convFn converterFn) (r rangeExpr, err error) {
	parts := strings.Split(expr, "-")
	if len(parts) != rangeSplitSize {
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
	if len(parts) != intervalSplitSize {
		return i, ErrMalformedField
	}

	i.incr, err = strconv.Atoi(parts[1])
	if err != nil {
		return i, err
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

func (notSpecifiedExpr) NearestCandidate(_ time.Time, other int, _ bool) (int, result) {
	return other, hit
}

func (notSpecifiedExpr) SubsetOf(_, _ int) bool {
	// A field not specified is always contained by any range of values.
	return true
}

// unitExpr is an expression field that represents a single possible value.
type unitExpr int

func (u unitExpr) NearestCandidate(_ time.Time, other int, forwards bool) (int, result) {
	value := int(u)

	switch {
	case (forwards && value < other) || (!forwards && value > other):
		return value, miss
	case value == other:
		return value, hit
	default:
		return value, inRange
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

func (r rangeExpr) NearestCandidate(t time.Time, other int, forwards bool) (int, result) {
	from, to := r.from, r.to
	if !forwards {
		from, to = to, from
	}

	// Check if the value lies after the maximum of the range which would mean a
	// miss and the search must move to the next candidate.
	if _, res := to.NearestCandidate(t, other, forwards); res == miss {
		return other, miss
	}

	// Check if the value lies before the minimum of the range which means the
	// nearest candidate is the beginning of the range.
	value, res := from.NearestCandidate(t, other, forwards)
	if res == inRange {
		return value, inRange
	}

	// The value is inside the range.
	return other, hit
}

func (r rangeExpr) SubsetOf(min, max int) bool {
	return r.from.SubsetOf(min, max) && r.to.SubsetOf(min, max)
}

type intervalExpr struct {
	rge  rangeExpr
	incr int
}

func (i intervalExpr) NearestCandidate(t time.Time, other int, forwards bool) (int, result) {
	// Get the beginning of the range.
	from, _ := i.rge.from.NearestCandidate(t, other, forwards)

	remainder := (other - from) % i.incr
	value := other
	if remainder > 0 {
		value += i.incr - remainder
	}
	if !forwards {
		value -= i.incr
	}

	_, res := i.rge.NearestCandidate(t, value, forwards)
	switch {
	case res == hit && other == value:
		return value, hit
	case res == inRange || res == hit:
		return max(from, value), inRange
	default:
		return value, miss
	}
}

func (i intervalExpr) SubsetOf(min, max int) bool {
	return i.rge.SubsetOf(min, max)
}

// nthLastDayOfMonthExpr is a specialized expression field to determine the last
// day of a month.
type nthLastDayOfMonthExpr struct {
	nthLast int
}

func (e nthLastDayOfMonthExpr) NearestCandidate(t time.Time, other int, forwards bool) (int, result) {
	value := findLastDayOfMonth(t).Day() - e.nthLast
	_, res := unitExpr(value).NearestCandidate(t, other, forwards)
	return value, res
}

func (e nthLastDayOfMonthExpr) SubsetOf(min, max int) bool {
	return e.nthLast+1 >= min && max-e.nthLast >= min
}

// lastWeekDayOfMonthExpr is a specialized expression field to determine which
// day of the month corresponds to the last occurrence of a week day.
type lastWeekDayOfMonthExpr struct {
	weekday time.Weekday
}

func (e lastWeekDayOfMonthExpr) NearestCandidate(t time.Time, _ int, forwards bool) (int, result) {
	value := e.valueFor(t)
	_, direction := unitExpr(value).NearestCandidate(t, t.Day(), forwards)
	return value, direction
}

func (e lastWeekDayOfMonthExpr) valueFor(t time.Time) int {
	lastDayOfMonth := findLastDayOfMonth(t)
	diff := int(lastDayOfMonth.Weekday() - e.weekday)
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

type nthWeekdayOfMonthExpr struct {
	weekday time.Weekday
	nth     int
}

func (e nthWeekdayOfMonthExpr) NearestCandidate(t time.Time, _ int, forwards bool) (int, result) {
	firstDayOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())

	diff := int(firstDayOfMonth.Weekday() - e.weekday)
	if diff <= 0 {
		// Add a week length to get a positive value. Note that here we also
		// increment foor the zero value.
		diff += int(time.Saturday) + 1
	}

	diff = -diff + e.nth*(int(time.Saturday+1))

	value := firstDayOfMonth.AddDate(0, 0, diff).Day()
	_, direction := unitExpr(value).NearestCandidate(t, t.Day(), forwards)
	return value, direction
}

func (e nthWeekdayOfMonthExpr) SubsetOf(min, max int) bool {
	return e.nth >= 1 && e.nth <= 5 && int(e.weekday) >= min && int(e.weekday) <= max
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
	if parts := strings.Split(value, "#"); len(parts) == nthSplitSize {
		weekday, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		nth, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}

		return nthWeekdayOfMonthExpr{weekday: time.Weekday(weekday), nth: nth}, nil
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

type result int

const (
	miss result = iota
	hit
	inRange
)

// timeSet represents a set of possible values for a time unit.
type timeSet interface {
	// NearestCandidate returns the value and its direction compared the given
	// time and value.
	NearestCandidate(time.Time, int, bool) (int, result)

	// SubsetOf returns true if the time set is included in the range [min,
	// max].
	SubsetOf(min, max int) bool
}

type converterFn func(string) (timeSet, error)
