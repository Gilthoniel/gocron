# gocron

This library is work in progress.

| Field name   | Allowed values | Allowed special characters |
| ------------ | -------------- | -------------------------- |
| Seconds      | 0-59           | , - * /                    |
| Minutes      | 0-59           | , - * /                    |
| Hours        | 0-59           | , - * /                    |
| Day of month | 1-31           | , - * / L                  |
| Month        | 1-12           | , - * /                    |
| Day of week  | 0-6 or SUN-SAT | , - * / L                  |

## Reference

This library implements the reference found in [Wikipedia](https://en.wikipedia.org/wiki/Cron) 
including the non-standard characters.
