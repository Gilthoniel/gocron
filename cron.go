package gocron

import "time"

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

func (s Schedule) Next() time.Time {
	return s.NextAfter(time.Now())
}

func (s Schedule) NextAfter(after time.Time) (next time.Time) {
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

type TimeUnit interface {
	Next(next time.Time) (time.Time, bool)
}
