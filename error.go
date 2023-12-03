package gocron

import (
	"errors"
	"fmt"
)

var (
	ErrMultipleNotSpecified = errors.New("only one `?` is supported")
)

type TimeUnitError struct {
	inner error
	kind  TimeUnitKind
}

func newTimeUnitErr(kind TimeUnitKind, inner error) TimeUnitError {
	return TimeUnitError{inner: inner, kind: kind}
}

func (e TimeUnitError) Error() string {
	return fmt.Sprintf("time unit `%s` malformed: %s", e.kind, e.inner)
}

func (e TimeUnitError) Is(err error) bool {
	tue, ok := err.(TimeUnitError)
	return ok && tue.kind == e.kind && errors.Is(tue.inner, e.inner)
}

type TimeUnitKind int

const (
	Seconds TimeUnitKind = iota
	Minutes
	Hours
	Days
	Months
	WeekDays
)

var kinds = []string{"seconds", "minutes", "hours", "days", "months", "week days"}

func (k TimeUnitKind) String() string {
	return kinds[k]
}
