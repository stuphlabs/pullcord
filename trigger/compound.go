package trigger

import (
	//github.com/stuphlabs/pullcord
)

// CompoundTrigger is a TriggerHandler that allows more than one trigger to be
// fired off at a time. It any trigger returns an error, it isn't guaranteed
// that all triggers will fire.
type CompoundTrigger struct {
	Triggers []TriggerHandler
}

// TriggerString implements the required string-based triggering function to
// make CompoundTrigger a valid TriggerHandler implementation.
func (ct *CompoundTrigger) Trigger() error {
	for _, t := range ct.Triggers {
		if err := t.Trigger(); err != nil {
			return err
		}
	}
	return nil
}

