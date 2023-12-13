package gocron

import (
	"errors"
	"testing"
)

func TestParser_Parse_refusesTwiceNotSpecified(t *testing.T) {
	_, err := defaultParser.Parse("* * * ? * ?")
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

	parser := defaultParser

	for _, v := range vectors {
		t.Run(v, func(t *testing.T) {
			_, err := parser.Parse(v)
			requireErrorIs(t, err, ErrValueOutsideRange)
		})
	}
}

func TestParser_Parse_abortsOnMalformedSeconds(t *testing.T) {
	_, err := defaultParser.Parse("a * * * * *")

	var e TimeUnitError
	requireErrorAs(t, err, &e)
	requireSameKind(t, e.kind, Seconds)
}

func TestParser_Parse_abortsOnMalformedMinutes(t *testing.T) {
	_, err := defaultParser.Parse("* a * * * *")

	var e TimeUnitError
	requireErrorAs(t, err, &e)
	requireSameKind(t, e.kind, Minutes)
}

func TestParser_Parse_abortsOnMalformedHours(t *testing.T) {
	_, err := defaultParser.Parse("* * a * * *")

	var e TimeUnitError
	requireErrorAs(t, err, &e)
	requireSameKind(t, e.kind, Hours)
}

func TestParser_Parse_abortsOnMalformedDays(t *testing.T) {
	_, err := defaultParser.Parse("* * * a * *")

	var e TimeUnitError
	requireErrorAs(t, err, &e)
	requireSameKind(t, e.kind, Days)
}

func TestParser_Parse_abortsOnMalformedMonths(t *testing.T) {
	_, err := defaultParser.Parse("* * * * a *")

	var e TimeUnitError
	requireErrorAs(t, err, &e)
	requireSameKind(t, e.kind, Months)
}

func TestParser_Parse_abortsOnMalformedWeekDays(t *testing.T) {
	_, err := defaultParser.Parse("* * * * * a")

	var e TimeUnitError
	requireErrorAs(t, err, &e)
	requireSameKind(t, e.kind, WeekDays)
}

func TestParser_Parse_abortsOnTooBigLastValue(t *testing.T) {
	_, err := defaultParser.Parse("* * * L-32 * *")
	requireErrorIs(t, err, ErrValueOutsideRange)
}

func TestParser_Parse_abortsOnTooSmallLastValue(t *testing.T) {
	_, err := defaultParser.Parse("* * * L-0 * *")
	requireErrorIs(t, err, ErrValueOutsideRange)
}

// --- Utilities

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
