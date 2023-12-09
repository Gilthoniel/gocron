package gocron

import (
	"errors"
	"slices"
	"sort"
	"strconv"
	"testing"
)

func TestParser_Parse_refusesTwiceNotSpecified(t *testing.T) {
	_, err := Parser{}.Parse("* * * ? * ?")
	requireErrorIs(t, err, ErrMultipleNotSpecified)
}

func TestParser_Parse_abortsOnValuesOutsideRange(t *testing.T) {
	vectors := []string{
		"0-60 * * * * *",
		"60 * * * * *",
		"* 0-60 * * * *",
		"* 60 * * * *",
		"* * 0-24 * * *",
		"* * 24 * * *",
		"* * * 0-31 * *",
		"* * * 0 * *",
		"* * * 32 * *",
		"* * * * 0-12 *",
		"* * * * 13 *",
		"* * * * * 0-7",
		"* * * * * 7",
	}

	parser := Parser{}

	for _, v := range vectors {
		t.Run(v, func(t *testing.T) {
			_, err := parser.Parse(v)
			requireErrorIs(t, err, ErrValueOutsideRange)
		})
	}
}

func TestParser_Parse_abortsOnMalformedSeconds(t *testing.T) {
	_, err := Parser{}.Parse("a * * * * *")

	var e TimeUnitError
	requireErrorAs(t, err, &e)
	requireSameKind(t, e.kind, Seconds)
}

func TestParser_Parse_abortsOnMalformedMinutes(t *testing.T) {
	_, err := Parser{}.Parse("* a * * * *")

	var e TimeUnitError
	requireErrorAs(t, err, &e)
	requireSameKind(t, e.kind, Minutes)
}

func TestParser_Parse_abortsOnMalformedHours(t *testing.T) {
	_, err := Parser{}.Parse("* * a * * *")

	var e TimeUnitError
	requireErrorAs(t, err, &e)
	requireSameKind(t, e.kind, Hours)
}

func TestParser_Parse_abortsOnMalformedDays(t *testing.T) {
	_, err := Parser{}.Parse("* * * a * *")

	var e TimeUnitError
	requireErrorAs(t, err, &e)
	requireSameKind(t, e.kind, Days)
}

func TestParser_Parse_abortsOnMalformedMonths(t *testing.T) {
	_, err := Parser{}.Parse("* * * * a *")

	var e TimeUnitError
	requireErrorAs(t, err, &e)
	requireSameKind(t, e.kind, Months)
}

func TestParser_Parse_abortsOnMalformedWeekDays(t *testing.T) {
	_, err := Parser{}.Parse("* * * * * a")

	var e TimeUnitError
	requireErrorAs(t, err, &e)
	requireSameKind(t, e.kind, WeekDays)
}

func TestParser_Parse_abortsOnTooBigLastValue(t *testing.T) {
	_, err := Parser{}.Parse("* * * L-32 * *")
	requireErrorIs(t, err, ErrValueOutsideRange)
}

func TestParser_Parse_abortsOnTooSmallLastValue(t *testing.T) {
	_, err := Parser{}.Parse("* * * L-0 * *")
	requireErrorIs(t, err, ErrValueOutsideRange)
}

func TestParser_sortableUnit(t *testing.T) {
	vectors := []struct {
		unsorted []timeSet
		expect   []timeSet
	}{
		{[]timeSet{unit(30), unit(15)}, []timeSet{unit(15), unit(30)}},
		{[]timeSet{unit(5), rge(1, 10)}, []timeSet{rge(1, 10), unit(5)}},
		{[]timeSet{rge(5, 10), unit(2)}, []timeSet{unit(2), rge(5, 10)}},
		{[]timeSet{interval(5, 10), unit(2)}, []timeSet{unit(2), interval(5, 10)}},
		{[]timeSet{unit(5), interval(1, 10)}, []timeSet{interval(1, 10), unit(5)}},
		{[]timeSet{interval(5, 10), rge(2, 15)}, []timeSet{rge(2, 15), interval(5, 10)}},
		{[]timeSet{rge(5, 10), interval(2, 15)}, []timeSet{interval(2, 15), rge(5, 10)}},
		{[]timeSet{interval(5, 10), interval(2, 15)}, []timeSet{interval(2, 15), interval(5, 10)}},
	}

	for i, vector := range vectors {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			sort.Sort(sortableUnit(vector.unsorted))
			if !slices.Equal(vector.unsorted, vector.expect) {
				t.Fatal(i, "not sorted", vector.unsorted)
			}
		})
	}
}

// --- Utilities

func unit(value int) unitExpr {
	return unitExpr(value)
}

func rge(from, to int) rangeExpr {
	return rangeExpr{from: unit(from), to: unit(to)}
}

func interval(from, to int) intervalExpr {
	return intervalExpr{rge: rangeExpr{unitExpr(from), unitExpr(to)}, incr: 1}
}

func requireErrorIs(t testing.TB, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("expected error: %#v, found %#v", target, err)
	}
}

func requireErrorAs(t testing.TB, err error, target any) {
	t.Helper()
	if !errors.As(err, target) {
		t.Fatalf("expected error of kind: %T, found %#v", target, err)
	}
}

func requireSameKind(t testing.TB, a, b TimeUnitKind) {
	t.Helper()
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}
