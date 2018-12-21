package trigger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/proidiot/gone/log"
	"github.com/stuphlabs/pullcord/config"
)

// DelayTrigger is a Triggerrer that delays the execution of another
// trigger for at least a minimum amount of time after the most recent request.
// The obvious analogy would be a screen saver, which will start after a
// certain period has elapsed, but the timer is reset quite often.
type DelayTrigger struct {
	DelayedTrigger Triggerrer
	Delay          time.Duration
	c              chan<- interface{}
}

func init() {
	config.MustRegisterResourceType(
		"delaytrigger",
		func() json.Unmarshaler {
			return new(DelayTrigger)
		},
	)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (d *DelayTrigger) UnmarshalJSON(input []byte) error {
	var t struct {
		DelayedTrigger config.Resource
		Delay          string
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	dt := t.DelayedTrigger.Unmarshalled
	switch dt := dt.(type) {
	case Triggerrer:
		d.DelayedTrigger = dt
	default:
		_ = log.Err(
			fmt.Sprintf(
				"Registry value is not a Trigger: %s",
				dt,
			),
		)
		return config.UnexpectedResourceType
	}

	dp, e := time.ParseDuration(t.Delay)
	if e != nil {
		return e
	}

	d.Delay = dp
	return nil
}

// NewDelayTrigger initializes a DelayTrigger. It might not be strictly
// necessary anymore.
func NewDelayTrigger(
	delayedTrigger Triggerrer,
	delay time.Duration,
) *DelayTrigger {
	return &DelayTrigger{
		delayedTrigger,
		delay,
		nil,
	}
}

func delaytrigger(
	tr Triggerrer,
	dla time.Duration,
	ac <-chan interface{},
) {
	tmr := time.NewTimer(dla)
	for {
		select {
		case _, ok := <-ac:
			if tmr != nil && !tmr.Stop() {
				<-tmr.C
			}
			if !ok {
				return
			}

			tmr.Reset(dla)
		case <-tmr.C:
			if err := tr.Trigger(); err != nil {
				_ = log.Err(
					fmt.Sprintf(
						"delaytrigger received an"+
							" error: %#v",
						err,
					),
				)
			}
			tmr.Stop()
		}
	}
}

// Trigger sets or resets the delay after which it will execute the child
// trigger. The child trigger will be executed no sooner than the delay time
// after any particular call, but subsequent calls may extend that time out
// further (possibly indefinitely).
func (d *DelayTrigger) Trigger() error {
	if d.c == nil {
		fc := make(chan interface{})
		d.c = fc

		go delaytrigger(d.DelayedTrigger, d.Delay, fc)
	} else {
		d.c <- nil
	}

	return nil
}
