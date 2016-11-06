package trigger

import (
	"fmt"
	//"github.com/stuphlabs/pullcord"
	"time"
)

// DelayTrigger is a TriggerHandler that delays the execution of another
// trigger for at least a minimum amount of time after the most recent request.
// The obvious analogy would be a screen saver, which will start after a
// certain period has elapsed, but the timer is reset quite often.
type DelayTrigger struct {
	Trigger TriggerHandler
	Delay time.Duration
	c chan<- string
}

func delaytrigger(tr TriggerHandler, dla time.Duration, ac <-chan string) {
	var tmr *time.Timer
	var arg string
	for {
		var c *<-chan time.Time
		if tmr == nil {
			c = nil
		} else {
			c = &tmr.C
		}
		select {
		case narg, ok := <-ac:
			if tmr != nil && !tmr.Stop() {
				<-tmr.C
			}
			if !ok {
				return
			} else {
				arg = narg
				tmr.Reset(dla)
			}
		case <-*c:
			if err := tr.TriggerString(arg); err != nil {
				log().Err(
					fmt.Sprintf(
						"delaytrigger received an",
						" error: %v",
						err,
					),
				)
			}
			tmr = nil
		}
	}
}

// TriggerString implements the required string-based triggering function to
// make DelayTrigger a valid TriggerHandler implementation. This function
// effectively cancels any previous trigger and replaces it with a call using
// only this most recent string value.
func (dt *DelayTrigger) TriggerString(arg string) error {
	if dt.c == nil {
		fc := make(chan string)
		dt.c = fc

		go delaytrigger(dt.Trigger, dt.Delay, fc)
	}

	dt.c <-arg

	return nil
}

