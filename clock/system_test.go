package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClockReturnsCurrentTime(t *testing.T) {
	c := SystemClock{}
	now := time.Now()
	system := c.GetCurrentTime()

	assert.True(t, now.Before(system))
}
