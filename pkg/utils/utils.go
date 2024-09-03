package utils

import "time"

// formatTimestamp formats a given timestamp into a UTC format time and returns the string.
func FormatTimestamp(timestamp string) (string, error) {
	t, err := time.Parse("2006-01-02T15:04:05Z", timestamp)

	if err != nil {
		return "", err
	}

	return t.Format("01-02-2006 15:04 UTC"), nil
}

func DifferenceOfSlices(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
