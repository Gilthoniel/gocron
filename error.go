package gocron

import (
	"errors"
	"fmt"
)

var (
	ErrMalformedExpression  = errors.New("expression is malformed")
	ErrMalformedField       = errors.New("unexpected field value")
	ErrMultipleNotSpecified = errors.New("only one `?` is supported")
	ErrValueOutsideRange    = errors.New("values are outside the supported range")
)

// TimeUnitError is an error returned when a time unit of a Cron expression is
// malformed.
type TimeUnitError struct {
	inner error
	kind  TimeUnitKind
}

func newTimeUnitErr(kind TimeUnitKind, inner error) TimeUnitError {
	return TimeUnitError{inner: inner, kind: kind}
}

// Kind returns the time unit kind of the error (e.g. seconds).
func (e TimeUnitError) Kind() TimeUnitKind {
	return e.kind
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
	Years
)

var kinds = []string{"seconds", "minutes", "hours", "days", "months", "week days", "years"}

func (k TimeUnitKind) String() string {
	return kinds[int(k)%len(kinds)]
}
