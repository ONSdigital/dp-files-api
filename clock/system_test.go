package clock

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClockReturnsCurrentTime(t *testing.T) {
	c := SystemClock{}
	now := time.Now()
	system := c.GetCurrentTime()
	
	assert.True(t, now.Before(system))
}
