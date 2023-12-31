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

func TestSchedule_Upcoming_returnsNextSecondsInProperOrder(t *testing.T) {
	iter := MustParse("0,45,15,30 * * * * *").Upcoming(time.Date(2000, time.March, 15, 12, 5, 1, 0, time.UTC))
	expects := []string{
		"2000-03-15 12:05:15 +0000 UTC",
		"2000-03-15 12:05:30 +0000 UTC",
		"2000-03-15 12:05:45 +0000 UTC",
		"2000-03-15 12:06:00 +0000 UTC",
		"2000-03-15 12:06:15 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

func TestSchedule_Upcoming_returnsNextNthLastDayOfMonth(t *testing.T) {
	iter := MustParse("0 0 0 L-31 * *").Upcoming(time.Date(2000, time.March, 15, 12, 5, 1, 0, time.UTC))
	expects := []string{
		"2000-05-01 00:00:00 +0000 UTC",
		"2000-07-01 00:00:00 +0000 UTC",
		"2000-08-01 00:00:00 +0000 UTC",
		"2000-10-01 00:00:00 +0000 UTC",
		"2000-12-01 00:00:00 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

func TestSchedule_Upcoming_returnsNextNthWeekdayOfMonth(t *testing.T) {
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

	testIterator(t, iter, expects)
}

func TestSchedule_Upcoming_returnsNextYear(t *testing.T) {
	iter := MustParse("0 0 0 1 6 ? 2010-2012").Upcoming(time.Date(2000, time.March, 15, 12, 5, 1, 0, time.UTC))
	expects := []string{
		"2010-06-01 00:00:00 +0000 UTC",
		"2011-06-01 00:00:00 +0000 UTC",
		"2012-06-01 00:00:00 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

func TestSchedule_Next_abortsExpressionWhichIsImpossible(t *testing.T) {
	schedule := MustParse("* * * 31 2 ?")

	next := schedule.Next(time.Now())
	if !next.IsZero() {
		t.Fatal("it should return the zero values")
	}
}

func TestSchedule_Preceding_returnsPreviousSeconds(t *testing.T) {
	iter := MustParse("15,5 * * * * *").Preceding(time.Date(2000, time.March, 15, 12, 5, 10, 0, time.UTC))
	expects := []string{
		"2000-03-15 12:05:05 +0000 UTC",
		"2000-03-15 12:04:15 +0000 UTC",
		"2000-03-15 12:04:05 +0000 UTC",
		"2000-03-15 12:03:15 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

func TestSchedule_Preceding_returnsPreviousMinutes(t *testing.T) {
	iter := MustParse("0 10,50 * * * *").Preceding(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	expects := []string{
		"2000-03-15 12:10:00 +0000 UTC",
		"2000-03-15 11:50:00 +0000 UTC",
		"2000-03-15 11:10:00 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

func TestSchedule_Preceding_returnsPreviousHours(t *testing.T) {
	iter := MustParse("0 0 5,17 * * *").Preceding(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	expects := []string{
		"2000-03-15 05:00:00 +0000 UTC",
		"2000-03-14 17:00:00 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

func TestSchedule_Preceding_returnsPreviousDays(t *testing.T) {
	iter := MustParse("30 30 12 28,31 * *").Preceding(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	expects := []string{
		"2000-02-28 12:30:30 +0000 UTC",
		"2000-01-31 12:30:30 +0000 UTC",
		"2000-01-28 12:30:30 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

func TestSchedule_Preceding_returnsPreviousMonths(t *testing.T) {
	iter := MustParse("0 0 0 1 2,6 ?").Preceding(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	expects := []string{
		"2000-02-01 00:00:00 +0000 UTC",
		"1999-06-01 00:00:00 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

func TestSchedule_Preceding_returnsPreviousWeekDays(t *testing.T) {
	iter := MustParse("0 0 0 ? * MON").Preceding(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	expects := []string{
		"2000-03-13 00:00:00 +0000 UTC",
		"2000-03-06 00:00:00 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

func TestSchedule_Preceding_returnsPreviousYear(t *testing.T) {
	iter := MustParse("0 0 0 1 6 ? 1900/5").Preceding(time.Date(2003, time.March, 15, 12, 5, 1, 0, time.UTC))
	expects := []string{
		"2000-06-01 00:00:00 +0000 UTC",
		"1995-06-01 00:00:00 +0000 UTC",
		"1990-06-01 00:00:00 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

func TestSchedule_Preceding_returnsPreviousYearInList(t *testing.T) {
	iter := MustParse("30 30 12 1 6 ? 1901,1902").Preceding(time.Date(2003, time.March, 15, 12, 5, 1, 0, time.UTC))
	expects := []string{
		"1902-06-01 12:30:30 +0000 UTC",
		"1901-06-01 12:30:30 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

func TestSchedule_Preceding_returnsPreviousInRange(t *testing.T) {
	iter := MustParse("0-1 * * * * *").Preceding(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	expects := []string{
		"2000-03-15 12:29:01 +0000 UTC",
		"2000-03-15 12:29:00 +0000 UTC",
		"2000-03-15 12:28:01 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

func TestSchedule_Preceding_returnsPreviousInInterval(t *testing.T) {
	iter := MustParse("30/15 * * * * *").Preceding(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	expects := []string{
		"2000-03-15 12:29:45 +0000 UTC",
		"2000-03-15 12:29:30 +0000 UTC",
		"2000-03-15 12:28:45 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

func TestSchedule_Preceding_returnsPreviousNthWeekday(t *testing.T) {
	iter := MustParse("0 0 0 ? * 6#1").Preceding(time.Date(2000, time.March, 15, 12, 30, 0, 0, time.UTC))
	expects := []string{
		"2000-03-04 00:00:00 +0000 UTC",
		"2000-02-05 00:00:00 +0000 UTC",
		"2000-01-01 00:00:00 +0000 UTC",
		"1999-12-04 00:00:00 +0000 UTC",
	}

	testIterator(t, iter, expects)
}

// --- Utilities

func requireTimeEqual(t testing.TB, value time.Time, expected string) {
	t.Helper()
	if value.String() != expected {
		t.Fatalf("%s != %s", value, expected)
	}
}

func testIterator(t *testing.T, iter *Iterator, expects []string) {
	for _, expect := range expects {
		t.Run(expect, func(t *testing.T) {
			requireTimeEqual(t, iter.Next(), expect)
		})
	}
}
