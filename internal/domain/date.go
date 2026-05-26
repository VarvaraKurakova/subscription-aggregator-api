package domain

import (
	"fmt"
	"time"
)

const MonthYearLayout = "01-2006"

func ParseMonthYear(value string) (time.Time, error) {
	parsed, err := time.Parse(MonthYearLayout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format %q, expected MM-YYYY: %w", value, err)
	}

	return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func FormatMonthYear(value time.Time) string {
	return value.Format(MonthYearLayout)
}

func MonthsBetweenInclusive(start, end time.Time) int {
	years := end.Year() - start.Year()
	months := int(end.Month()) - int(start.Month())

	return years*12 + months + 1
}

func MaxMonth(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}

	return b
}

func MinMonth(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}

	return b
}
