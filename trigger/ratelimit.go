package trigger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/proidiot/gone/log"
	"github.com/stuphlabs/pullcord/config"
)

// ErrRateLimitExceeded indicates that the trigger has been called more than
// the allowed number of times within the specified duration, and so the
// guarded trigger will not be called.
var ErrRateLimitExceeded = errors.New("Rate limit exceeded for trigger")

// RateLimitTrigger is a Triggerrer that will prevent a guarded trigger
// from being called more than a specified number of times over a specified
// duration.
type RateLimitTrigger struct {
	GuardedTrigger   Triggerrer
	MaxAllowed       uint
	Period           time.Duration
	previousTriggers []time.Time
}

func init() {
	config.MustRegisterResourceType(
		"ratelimittrigger",
		func() json.Unmarshaler {
			return new(RateLimitTrigger)
		},
	)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (r *RateLimitTrigger) UnmarshalJSON(input []byte) error {
	var t struct {
		GuardedTrigger config.Resource
		MaxAllowed     uint
		Period         string
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	gt := t.GuardedTrigger.Unmarshalled
	switch gt := gt.(type) {
	case Triggerrer:
		r.GuardedTrigger = gt
	default:
		_ = log.Err(
			fmt.Sprintf(
				"Registry value is not a Trigger: %s",
				gt,
			),
		)
		return config.UnexpectedResourceType
	}

	p, e := time.ParseDuration(t.Period)
	if e != nil {
		return e
	}

	r.Period = p

	r.MaxAllowed = t.MaxAllowed

	return nil
}

// NewRateLimitTrigger initializes a RateLimitTrigger. It may no longer be
// strictly necessary.
func NewRateLimitTrigger(
	guardedTrigger Triggerrer,
	maxAllowed uint,
	period time.Duration,
) *RateLimitTrigger {
	return &RateLimitTrigger{
		guardedTrigger,
		maxAllowed,
		period,
		nil,
	}
}

// Trigger executes its guarded trigger if and only if it has not be called more
// than the allowed number of times within the specified rolling window of time.
// If the rate limit is exceeded, ErrRateLimitExceeded will be returned, and
// the guarded trigger will not be called.
func (r *RateLimitTrigger) Trigger() error {
	now := time.Now()
	_ = log.Debug("rate limit trigger initiated")

	if r.previousTriggers != nil {
		_ = log.Debug("determine if rate limit has been exceeded")
		for len(
			r.previousTriggers,
		) > 0 && now.After(r.previousTriggers[0].Add(r.Period)) {
			r.previousTriggers = r.previousTriggers[1:]
		}

		if uint(len(r.previousTriggers)) >= r.MaxAllowed {
			_ = log.Debug("rate limit has been exceeded")
			return ErrRateLimitExceeded
		}
	} else {
		_ = log.Debug("first rate limited trigger")
		r.previousTriggers = make([]time.Time, r.MaxAllowed)
	}

	r.previousTriggers = append(r.previousTriggers, now)

	_ = log.Debug("rate limit not exceeded, cascading the trigger")
	return r.GuardedTrigger.Trigger()
}
