package trigger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stuphlabs/pullcord/config"
	"time"
)

// DelayTrigger is a TriggerHandler that delays the execution of another
// trigger for at least a minimum amount of time after the most recent request.
// The obvious analogy would be a screen saver, which will start after a
// certain period has elapsed, but the timer is reset quite often.
type DelayTrigger struct {
	DelayedTrigger TriggerHandler
	Delay time.Duration
	c chan<- interface{}
}

func init() {
	config.RegisterResourceType(
		"delaytrigger",
		func() json.Unmarshaler {
			return new(DelayTrigger)
		},
	)
}

func (d *DelayTrigger) UnmarshalJSON(input []byte) (error) {
	var t struct {
		DelayedTrigger config.Resource
		Delay string
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		return e
	}

	dt := t.DelayedTrigger.Unmarshaled
	switch dt := dt.(type) {
	case TriggerHandler:
		d.DelayedTrigger = dt
	default:
		log().Err(
			fmt.Sprintf(
				"Registry value is not a Trigger: %s",
				dt,
			),
		)
		return config.UnexpectedResourceType
	}

	if dp, e := time.ParseDuration(t.Delay); e != nil {
		return e
	} else {
		d.Delay = dp
	}

	return nil
}

func NewDelayTrigger(
	delayedTrigger TriggerHandler,
	delay time.Duration,
) (*DelayTrigger) {
	return &DelayTrigger{
		delayedTrigger,
		delay,
		nil,
	}
}

func delaytrigger(
	tr TriggerHandler,
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
			} else {
				tmr.Reset(dla)
			}
		case <-tmr.C:
			if err := tr.Trigger(); err != nil {
				log().Err(
					fmt.Sprintf(
						"delaytrigger received an",
						" error: %v",
						err,
					),
				)
			}
			tmr.Stop()
		}
	}
}

// TriggerString implements the required string-based triggering function to
// make DelayTrigger a valid TriggerHandler implementation. This function
// effectively cancels any previous trigger and replaces it with a call using
// only this most recent string value.
func (dt *DelayTrigger) Trigger() error {
	if dt.c == nil {
		fc := make(chan interface{})
		dt.c = fc

		go delaytrigger(dt.DelayedTrigger, dt.Delay, fc)
	} else {
		dt.c <-nil
	}

	return nil
}

