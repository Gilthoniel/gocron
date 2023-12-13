package gocron

import (
	"testing"
	"time"
)

func TestSchedule_Next(t *testing.T) {
	after := time.Date(2000, time.March, 15, 12, 5, 1, 0, time.UTC)
	vectors := []struct {
		expr   string
		expect time.Time
	}{
		{"* * * * * *", time.Date(2000, time.March, 15, 12, 5, 2, 0, time.UTC)},
		{"30 * * * * *", time.Date(2000, time.March, 15, 12, 5, 30, 0, time.UTC)},
		{"15,30 * * * * *", time.Date(2000, time.March, 15, 12, 5, 15, 0, time.UTC)},
		{"0-2 * * * * *", time.Date(2000, time.March, 15, 12, 5, 2, 0, time.UTC)},
		{"*/15 * * * * *", time.Date(2000, time.March, 15, 12, 5, 15, 0, time.UTC)},
		{"50/10 * * * * *", time.Date(2000, time.March, 15, 12, 5, 50, 0, time.UTC)},
		{"30-59/5 * * * * *", time.Date(2000, time.March, 15, 12, 5, 30, 0, time.UTC)},
		{"0-1,30-40 * * * * *", time.Date(2000, time.March, 15, 12, 5, 30, 0, time.UTC)},
		{"0 * * * * *", time.Date(2000, time.March, 15, 12, 6, 0, 0, time.UTC)},
		{"* 10 * * * *", time.Date(2000, time.March, 15, 12, 10, 0, 0, time.UTC)},
		{"* 0-10 * * * *", time.Date(2000, time.March, 15, 12, 5, 2, 0, time.UTC)},
		{"* 1,2 * * * *", time.Date(2000, time.March, 15, 13, 1, 0, 0, time.UTC)},
		{"* 0/10 * * * *", time.Date(2000, time.March, 15, 12, 10, 0, 0, time.UTC)},
		{"* * 2 * * *", time.Date(2000, time.March, 16, 2, 0, 0, 0, time.UTC)},
		{"* * 10-13 * * *", time.Date(2000, time.March, 15, 12, 5, 2, 0, time.UTC)},
		{"* * 2,13 * * *", time.Date(2000, time.March, 15, 13, 0, 0, 0, time.UTC)},
		{"* * 0/10 * * *", time.Date(2000, time.March, 15, 20, 0, 0, 0, time.UTC)},
		{"* * * 31 4-7 *", time.Date(2000, time.May, 31, 0, 0, 0, 0, time.UTC)},
		{"* * * 14 3,5 *", time.Date(2000, time.May, 14, 0, 0, 0, 0, time.UTC)},
		{"* * * 15/1 * *", time.Date(2000, time.March, 15, 12, 5, 2, 0, time.UTC)},
		{"* * * 1-5/1 * *", time.Date(2000, time.April, 1, 0, 0, 0, 0, time.UTC)},
		{"* * * * 1-2 *", time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{"* * * L * *", time.Date(2000, time.March, 31, 0, 0, 0, 0, time.UTC)},
		{"* * * 1,L 4 *", time.Date(2000, time.April, 1, 0, 0, 0, 0, time.UTC)},
		{"* * * 20-L 4 *", time.Date(2000, time.April, 20, 0, 0, 0, 0, time.UTC)},
		{"* * * 1 12 SUN", time.Date(2002, time.December, 1, 0, 0, 0, 0, time.UTC)},
		{"* * * 1 12 0", time.Date(2002, time.December, 1, 0, 0, 0, 0, time.UTC)},
		{"* * * 1 12 SUN,FRI", time.Date(2000, time.December, 1, 0, 0, 0, 0, time.UTC)},
		{"* * * 1 12 SUN-FRI", time.Date(2000, time.December, 1, 0, 0, 0, 0, time.UTC)},
		{"* * * * * */2", time.Date(2000, time.March, 16, 0, 0, 0, 0, time.UTC)},
		{"* * * * * 1/2", time.Date(2000, time.March, 15, 12, 5, 2, 0, time.UTC)},
		{"* * * * * 6L", time.Date(2000, time.March, 25, 0, 0, 0, 0, time.UTC)},
		{"* * * * * 5L", time.Date(2000, time.March, 31, 0, 0, 0, 0, time.UTC)},
		{"* * * * * 4L", time.Date(2000, time.March, 30, 0, 0, 0, 0, time.UTC)},
		{"* * * 31 * L", time.Date(2001, time.March, 31, 0, 0, 0, 0, time.UTC)},
		{"* * * ? * 0L", time.Date(2000, time.March, 26, 0, 0, 0, 0, time.UTC)},
	}

	for _, v := range vectors {
		sched := MustParse(v.expr)

		t.Run(v.expr, func(t *testing.T) {
			after := sched.Next(after)
			if !v.expect.Equal(after) {
				t.Fatalf("%s", after)
			}
		})
	}
}

func TestSchedule_Next_returnsNextSecondsInProperOrder(t *testing.T) {
	iter := MustParse("0,45,15,30 * * * * *").Upcoming(time.Date(2000, time.March, 15, 12, 5, 1, 0, time.UTC))
	expects := []string{
		"2000-03-15 12:05:15 +0000 UTC",
		"2000-03-15 12:05:30 +0000 UTC",
		"2000-03-15 12:05:45 +0000 UTC",
		"2000-03-15 12:06:00 +0000 UTC",
		"2000-03-15 12:06:15 +0000 UTC",
	}

	for _, expect := range expects {
		t.Run(expect, func(t *testing.T) {
			requireTimeEqual(t, iter.Next(), expect)
		})
	}
}

func TestSchedule_Next_returnsNextNthLastDayOfMonth(t *testing.T) {
	iter := MustParse("0 0 0 L-31 * *").Upcoming(time.Date(2000, time.March, 15, 12, 5, 1, 0, time.UTC))
	expects := []string{
		"2000-05-01 00:00:00 +0000 UTC",
		"2000-07-01 00:00:00 +0000 UTC",
		"2000-08-01 00:00:00 +0000 UTC",
		"2000-10-01 00:00:00 +0000 UTC",
		"2000-12-01 00:00:00 +0000 UTC",
	}

	for _, expect := range expects {
		t.Run(expect, func(t *testing.T) {
			requireTimeEqual(t, iter.Next(), expect)
		})
	}
}

func TestSchedule_Next_returnsNextNthWeekdayOfMonth(t *testing.T) {
	iter := MustParse("0 0 0 ? * 0#3").Upcoming(time.Date(2000, time.March, 15, 12, 5, 1, 0, time.UTC))
	expects := []string{
		"2000-03-19 00:00:00 +0000 UTC",
		"2000-04-16 00:00:00 +0000 UTC",
		"2000-05-21 00:00:00 +0000 UTC",
		"2000-06-18 00:00:00 +0000 UTC",
		"2000-07-16 00:00:00 +0000 UTC",
		"2000-08-20 00:00:00 +0000 UTC",
		"2000-09-17 00:00:00 +0000 UTC",
		"2000-10-15 00:00:00 +0000 UTC",
		"2000-11-19 00:00:00 +0000 UTC",
		"2000-12-17 00:00:00 +0000 UTC",
		"2001-01-21 00:00:00 +0000 UTC",
	}

	for _, expect := range expects {
		t.Run(expect, func(t *testing.T) {
			requireTimeEqual(t, iter.Next(), expect)
		})
	}
}

func TestSchedule_Next_abortsExpressionWhichIsImpossible(t *testing.T) {
	schedule := MustParse("* * * 31 2 ?")

	next := schedule.Next(time.Now())
	if !next.IsZero() {
		t.Fatal("it should return the zero values")
	}
}

func TestSchedule_Previous_returnsPreviousSeconds(t *testing.T) {
	schedule := MustParse("15,5 * * * * *")

	prev := schedule.Previous(time.Date(2000, time.March, 15, 12, 5, 10, 0, time.UTC))
	requireTimeEqual(t, prev, "2000-03-15 12:05:05 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "2000-03-15 12:04:15 +0000 UTC")
}

func TestSchedule_Previous_returnsPreviousMinutes(t *testing.T) {
	schedule := MustParse("0 10,50 * * * *")

	prev := schedule.Previous(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	requireTimeEqual(t, prev, "2000-03-15 12:10:00 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "2000-03-15 11:50:00 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "2000-03-15 11:10:00 +0000 UTC")
}

func TestSchedule_Previous_returnsPreviousHours(t *testing.T) {
	schedule := MustParse("0 0 5,17 * * *")

	prev := schedule.Previous(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	requireTimeEqual(t, prev, "2000-03-15 05:00:00 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "2000-03-14 17:00:00 +0000 UTC")
}

func TestSchedule_Previous_returnsPreviousDays(t *testing.T) {
	schedule := MustParse("0 0 0 28,31 * *")

	prev := schedule.Previous(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	requireTimeEqual(t, prev, "2000-02-28 00:00:00 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "2000-01-31 00:00:00 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "2000-01-28 00:00:00 +0000 UTC")
}

func TestSchedule_Previous_returnsPreviousMonths(t *testing.T) {
	schedule := MustParse("0 0 0 1 2,6 ?")

	prev := schedule.Previous(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	requireTimeEqual(t, prev, "2000-02-01 00:00:00 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "1999-06-01 00:00:00 +0000 UTC")
}

func TestSchedule_Previous_returnsPreviousWeekDays(t *testing.T) {
	schedule := MustParse("0 0 0 ? * MON")

	prev := schedule.Previous(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	requireTimeEqual(t, prev, "2000-03-13 00:00:00 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "2000-03-06 00:00:00 +0000 UTC")
}

func TestSchedule_Previous_returnsPreviousInRange(t *testing.T) {
	schedule := MustParse("0-1 * * * * *")

	prev := schedule.Previous(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	requireTimeEqual(t, prev, "2000-03-15 12:29:01 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "2000-03-15 12:29:00 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "2000-03-15 12:28:01 +0000 UTC")
}

func TestSchedule_Previous_returnsPreviousInInterval(t *testing.T) {
	schedule := MustParse("30/15 * * * * *")

	prev := schedule.Previous(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	requireTimeEqual(t, prev, "2000-03-15 12:29:45 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "2000-03-15 12:29:30 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "2000-03-15 12:28:45 +0000 UTC")
}

func TestSchedule_Previous_returnsPreviousNthWeekday(t *testing.T) {
	schedule := MustParse("0 0 0 ? * 6#1")

	prev := schedule.Previous(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	requireTimeEqual(t, prev, "2000-03-04 00:00:00 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "2000-02-05 00:00:00 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "2000-01-01 00:00:00 +0000 UTC")

	prev = schedule.Previous(prev)
	requireTimeEqual(t, prev, "1999-12-04 00:00:00 +0000 UTC")
}

// --- Utilities

func requireTimeEqual(t testing.TB, value time.Time, expected string) {
	t.Helper()
	if value.String() != expected {
		t.Fatalf("%s != %s", value, expected)
	}
}
