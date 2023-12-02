package gocron

import "time"

type TimeUnit interface {
	Next(next time.Time) (time.Time, bool)
}

type Schedule struct {
	timeUnits []TimeUnit
}

func Parse(expression string) (Schedule, error) {
	return Parser{}.Parse(expression)
}

func Must(expression string) Schedule {
	schedule, err := Parse(expression)
	if err != nil {
		panic(err)
	}
	return schedule
}

// Next returns a time after the given argument, but never equals to it. A zero
// time is returned when none can be found.
func (s Schedule) Next(after time.Time) (next time.Time) {
	after = after.Truncate(1 * time.Second).Add(1 * time.Second)
	return s.nextAfter(after)
}

func (s Schedule) nextAfter(after time.Time) time.Time {
	var ok bool
	for _, unit := range s.timeUnits {
		after, ok = unit.Next(after)
		if !ok {
			return s.nextAfter(after)
		}
	}
	return after
}
