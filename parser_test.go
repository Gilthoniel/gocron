package gocron

import (
	"errors"
	"testing"
)

func TestParser_Parse_refusesTwiceNotSpecified(t *testing.T) {
	_, err := Parser{}.Parse("* * * ? * ?")
	requireErrorIs(t, err, ErrMultipleNotSpecified)
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
