package clock

import "time"

type SystemClock struct{}

func (c SystemClock) GetCurrentTime() time.Time {
	return time.Now()
}
