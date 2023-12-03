package gocron

import (
	"errors"
	"fmt"
)

var (
	ErrMultipleNotSpecified = errors.New("only one `?` is supported")
	ErrValueOutsideRange    = errors.New("values are outside the supported range")
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
	if ok {
		return tue.kind == e.kind && errors.Is(e.inner, tue.inner)
	}
	return errors.Is(e.inner, err)
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
