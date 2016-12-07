package trigger

import (
	"github.com/stretchr/testify/assert"
	//"github.com/stuphlabs/pullcord"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	testString := "testing"
	cth := &counterTriggerHandler{testString, 0}

	var rlt RateLimitTrigger
	rlt.GuardedTrigger = cth
	rlt.MaxAllowed = 1
	rlt.Period = time.Second
	err := rlt.TriggerString(testString)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), cth.count)

	err = rlt.TriggerString(testString)
	assert.Error(t, err)
	assert.Equal(t, RateLimitExceededError, err)
	assert.Equal(t, uint(1), cth.count)

	time.Sleep(time.Second)
	err = rlt.TriggerString(testString)
	assert.NoError(t, err)
	assert.Equal(t, uint(2), cth.count)
}

