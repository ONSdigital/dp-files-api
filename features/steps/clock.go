package steps

import "time"

type testClock struct{}

func (tt testClock) GetCurrentTime() time.Time {
	t, _ := time.Parse(time.RFC3339, "2021-10-19T09:30:30Z")
	return t
}
