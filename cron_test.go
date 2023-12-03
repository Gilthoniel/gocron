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
	}

	for _, v := range vectors {
		sched := Must(v.expr)

		t.Run(v.expr, func(t *testing.T) {
			after := sched.Next(after)
			if !v.expect.Equal(after) {
				t.Fatalf("%s", after)
			}
		})
	}
}
