package gocron_test

import (
	"fmt"
	"time"

	"github.com/Gilthoniel/gocron"
)

func ExampleSchedule_Next_everyFifteenSeconds() {
	schedule := gocron.Must("*/15 * * * ? *")

	next := schedule.Next(time.Date(2023, time.June, 4, 0, 0, 0, 0, time.UTC))
	fmt.Println(next)

	next = schedule.Next(next)
	fmt.Println(next)

	// Output:
	// 2023-06-04 00:00:15 +0000 UTC
	// 2023-06-04 00:00:30 +0000 UTC
}

func ExampleSchedule_Next_everyLastFridayOfTheMonthAtMidnight() {
	schedule := gocron.Must("0 0 0 ? * 5L")

	next := schedule.Next(time.Date(2023, time.June, 4, 0, 0, 0, 0, time.UTC))
	fmt.Println(next)

	next = schedule.Next(next)
	fmt.Println(next)

	// Output:
	// 2023-06-30 00:00:00 +0000 UTC
	// 2023-07-28 00:00:00 +0000 UTC
}