package trigger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stuphlabs/pullcord/config"
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

func init() {
	config.RegisterResourceType(
		"ratelimittrigger",
		func() json.Unmarshaler {
			return new(RateLimitTrigger)
		},
	)
}

func (r *RateLimitTrigger) UnmarshalJSON(input []byte) (error) {
	var t struct {
		GuardedTrigger config.Resource
		MaxAllowed uint
		Period string
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	gt := t.GuardedTrigger.Unmarshaled
	switch gt := gt.(type) {
	case TriggerHandler:
		r.GuardedTrigger = gt
	default:
		log().Err(
			fmt.Sprintf(
				"Registry value is not a Trigger: %s",
				gt,
			),
		)
		return config.UnexpectedResourceType
	}

	if p, e := time.ParseDuration(t.Period); e != nil {
		return e
	} else {
		r.Period = p
	}

	r.MaxAllowed = t.MaxAllowed

	return nil
}

func NewRateLimitTrigger(
	guardedTrigger TriggerHandler,
	maxAllowed uint,
	period time.Duration,
) (*RateLimitTrigger) {
	return &RateLimitTrigger{
		guardedTrigger,
		maxAllowed,
		period,
		nil,
	}
}

// TriggerString implements the required string-based triggering function to
// make RateLimitTrigger a valid TriggerHandler implementation. If the rate
// limit is exceeded, RateLimitExceededError will be returned, and the guarded
// trigger will not be called.
func (rlt *RateLimitTrigger) Trigger() error {
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

	return rlt.GuardedTrigger.Trigger()
}

