package steps

import "time"

type TestClock struct{}

func (tt TestClock) GetCurrentTime() time.Time {
	t, _ := time.Parse(time.RFC3339, "2021-10-19T09:30:30Z")
	return t
}
