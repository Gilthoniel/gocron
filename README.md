# gocron

[![Go Reference](https://pkg.go.dev/badge/github.com/Gilthoniel/gocron.svg)](https://pkg.go.dev/github.com/Gilthoniel/gocron)

The gocron package provides primitives to parse a Cron expression and iterate over the activation times.

### Install

```Shell
go get github.com/Gilthoniel/gocron
```

### Support

This package implements the specification found in [Wikipedia](https://en.wikipedia.org/wiki/Cron) and includes some advanced usages from the [Quartz scheduler](http://www.quartz-scheduler.org/documentation/quartz-2.3.0/tutorials/crontrigger.html).

| Field name   | Allowed values | Allowed special characters |
| ------------ | -------------- | -------------------------- |
| Seconds      | 0-59           | , - * /                    |
| Minutes      | 0-59           | , - * /                    |
| Hours        | 0-59           | , - * /                    |
| Day of month | 1-31           | , - * ? / L                |
| Month        | 1-12           | , - * /                    |
| Day of week  | 0-6 or SUN-SAT | , - * ? / L #              |
| Years        | 1-9999         | , - * /                    |

#### Special characters

* `*` can be used to select all values within a field (e.g. every second, every minute, ...)
* `?` can be used for no specific value in the day field or the week day field.
* `-` can be used to specify an inclusive range (e.g. `1-3` means values 1, 2, and 3).
* `,` can be used to list values (e.g. `1,2,3` means values 1, 2 and 3).
* `/` can be used to specify an interval (e.g. `1/5` means values 1, 6, 11, 16, etc...).
* `L` when used in the month field specifies the last day of the month and Saturday when used in the week day field. Using a digit before the character in the week day field specifies the nth last week day of the month (e.g. `1L` for the last Monday of the month). An offset can also be used for the month field (e.g. `L-2` for the second last day of the month).
* `#` can be used to specify the nth week day of the month (e.g. `6#3` for the third (`3`) Saturday (`6`) of the month).
