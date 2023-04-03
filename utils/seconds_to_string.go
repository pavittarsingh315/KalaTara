package utils

import "fmt"

func SecondsToString(seconds int64) string {
	minutes := seconds / 60
	hours := seconds / 3600
	days := seconds / (3600 * 24)

	type TimeUnit struct {
		Unit     int64
		Singular string
		Plural   string
	}

	// The order of these matters, because we want to display the largest unit first.
	timeUnits := []TimeUnit{
		{Unit: days, Singular: "day", Plural: "days"},
		{Unit: hours, Singular: "hour", Plural: "hours"},
		{Unit: minutes, Singular: "minute", Plural: "minutes"},
		{Unit: seconds, Singular: "second", Plural: "seconds"},
	}

	for _, timeUnit := range timeUnits {
		if timeUnit.Unit > 0 {
			var quantifier string
			if timeUnit.Unit > 1 {
				quantifier = timeUnit.Plural
			} else {
				quantifier = timeUnit.Singular
			}
			return fmt.Sprintf("%d %s", timeUnit.Unit, quantifier)
		}
	}

	return "0 seconds"
}
