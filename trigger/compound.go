package trigger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stuphlabs/pullcord/config"
)

// CompoundTrigger is a TriggerHandler that allows more than one trigger to be
// fired off at a time. It any trigger returns an error, it isn't guaranteed
// that all triggers will fire.
type CompoundTrigger struct {
	Triggers []TriggerHandler
}

func init() {
	config.RegisterResourceType(
		"compoundtrigger",
		func() json.Unmarshaler {
			return new(CompoundTrigger)
		},
	)
}

func (c *CompoundTrigger) UnmarshalJSON(input []byte) (error) {
	var t struct {
		Triggers []config.Resource
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		return e
	} else {
		c.Triggers = make([]TriggerHandler, len(t.Triggers))

		for _, i := range t.Triggers {
			th := i.Unmarshaled
			switch th := th.(type) {
			case TriggerHandler:
				c.Triggers = append(c.Triggers, th)
			default:
				log().Err(
					fmt.Sprintf(
						"Registry value is not a" +
						" RequestFilter: %s",
						th,
					),
				)
				return config.UnexpectedResourceType
			}
		}

		return nil
	}
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

