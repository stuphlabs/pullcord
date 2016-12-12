package trigger

import (
	"github.com/stretchr/testify/assert"
	//"github.com/stuphlabs/pullcord"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	cth := &counterTriggerHandler{}

	rlt := NewRateLimitTrigger(cth, 1, time.Second)

	err := rlt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 1, cth.count)

	err = rlt.Trigger()
	assert.Error(t, err)
	assert.Equal(t, RateLimitExceededError, err)
	assert.Equal(t, 1, cth.count)

	time.Sleep(time.Second)
	err = rlt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 2, cth.count)
}

