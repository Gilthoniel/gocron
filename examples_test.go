package gocron_test

import (
	"fmt"
	"time"

	"github.com/Gilthoniel/gocron"
)

func ExampleSchedule_Next_everyFifteenSeconds() {
	schedule := gocron.MustParse("*/15 * * * ? *")

	next := schedule.Next(time.Date(2023, time.June, 4, 0, 0, 0, 0, time.UTC))
	fmt.Println(next)

	next = schedule.Next(next)
	fmt.Println(next)

	// Output:
	// 2023-06-04 00:00:15 +0000 UTC
	// 2023-06-04 00:00:30 +0000 UTC
}

func ExampleSchedule_Next_usingTimezone() {
	schedule := gocron.MustParse("*/15 * * * * *")

	next := schedule.Next(time.Date(2023, time.June, 4, 0, 0, 0, 0, time.FixedZone("CEST", 120)))
	fmt.Println(next)

	next = schedule.Next(next)
	fmt.Println(next)

	// Output:
	// 2023-06-04 00:00:15 +0002 CEST
	// 2023-06-04 00:00:30 +0002 CEST
}

func ExampleSchedule_Next_everyLastFridayOfTheMonthAtMidnight() {
	schedule := gocron.MustParse("0 0 0 ? * 5L")

	next := schedule.Next(time.Date(2023, time.June, 4, 0, 0, 0, 0, time.UTC))
	fmt.Println(next)

	next = schedule.Next(next)
	fmt.Println(next)

	// Output:
	// 2023-06-30 00:00:00 +0000 UTC
	// 2023-07-28 00:00:00 +0000 UTC
}

func ExampleSchedule_Upcoming_everyLastSundayOfAprilAtThreePM() {
	schedule := gocron.MustParse("0 0 15 ? 4 0L")

	iter := schedule.Upcoming(time.Date(2023, time.June, 4, 0, 0, 0, 0, time.UTC))
	for i := 0; i < 5 && iter.HasNext(); i++ {
		next := iter.Next()
		fmt.Println(next)
	}

	// Output:
	// 2024-04-28 15:00:00 +0000 UTC
	// 2025-04-27 15:00:00 +0000 UTC
	// 2026-04-26 15:00:00 +0000 UTC
	// 2027-04-25 15:00:00 +0000 UTC
	// 2028-04-30 15:00:00 +0000 UTC
}

func ExampleSchedule_Upcoming_everySecondToLastDayOfEveryTwoMonths() {
	schedule := gocron.MustParse("0 0 0 L-2 */2 ?")

	iter := schedule.Upcoming(time.Date(2023, time.June, 4, 0, 0, 0, 0, time.UTC))
	for i := 0; i < 5 && iter.HasNext(); i++ {
		next := iter.Next()
		fmt.Println(next)
	}

	// Output:
	// 2023-07-30 00:00:00 +0000 UTC
	// 2023-09-29 00:00:00 +0000 UTC
	// 2023-11-29 00:00:00 +0000 UTC
	// 2024-01-30 00:00:00 +0000 UTC
	// 2024-03-30 00:00:00 +0000 UTC
}
