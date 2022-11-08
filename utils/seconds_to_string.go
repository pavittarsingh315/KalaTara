package utils

import "fmt"

func SecondsToString(seconds int64) string {
	minutes := seconds / 60
	hours := seconds / 3600
	days := seconds / (3600 * 24)

	var quantifier string
	if seconds < 60 {
		if seconds > 1 {
			quantifier = "seconds"
		} else {
			quantifier = "second"
		}
		return fmt.Sprintf("%d %s", seconds, quantifier)
	} else if minutes < 60 {
		if minutes > 1 {
			quantifier = "minutes"
		} else {
			quantifier = "minute"
		}
		return fmt.Sprintf("%d %s", minutes, quantifier)
	} else if hours < 24 {
		if hours > 1 {
			quantifier = "hours"
		} else {
			quantifier = "hour"
		}
		return fmt.Sprintf("%d %s", hours, quantifier)
	} else {
		if days > 1 {
			quantifier = "days"
		} else {
			quantifier = "day"
		}
		return fmt.Sprintf("%d %s", days, quantifier)
	}
}
