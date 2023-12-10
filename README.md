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
