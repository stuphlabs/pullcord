package trigger

import (
	"errors"
	//"github.com/stuphlabs/pullcord"
	"time"
)

// RateLimitExceededError indicates that the trigger has been called more than
// the allowed number of times within the specified duration, and so the
// guarded trigger will not be called.
var RateLimitExceededError = errors.New("Rate limit exceeded for trigger.")

// RateLimitTrigger is a TriggerHandler that will prevent a guarded trigger
// from being called more than a specified number of times over a specified
// duration.
type RateLimitTrigger struct {
	GuardedTrigger TriggerHandler
	MaxAllowed uint
	Period time.Duration
	previousTriggers []time.Time
}

// TriggerString implements the required string-based triggering function to
// make RateLimitTrigger a valid TriggerHandler implementation. If the rate
// limit is exceeded, RateLimitExceededError will be returned, and the guarded
// trigger will not be called.
func (rlt *RateLimitTrigger) TriggerString(arg string) error {
	now := time.Now()

	if rlt.previousTriggers != nil {
		for len(
			rlt.previousTriggers,
		) > 0 && now.After(rlt.previousTriggers[0].Add(rlt.Period)) {
			rlt.previousTriggers = rlt.previousTriggers[1:]
		}

		if uint(len(rlt.previousTriggers)) >= rlt.MaxAllowed {
			return RateLimitExceededError
		}
	} else {
		rlt.previousTriggers = make([]time.Time, rlt.MaxAllowed)
	}

	rlt.previousTriggers = append(rlt.previousTriggers, now)

	return rlt.GuardedTrigger.TriggerString(arg)
}

