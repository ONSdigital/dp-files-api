package clock

import "time"

type Clock interface {
	GetCurrentTime() time.Time
}
