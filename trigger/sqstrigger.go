package trigger

import (
	"errors"
	// "github.com/stuphlabs/pullcord"
)

// SqsTriggerrer is a work-in-progress that is intended to eventually send
// SQS messages as the result of a triggering event.
type SqsTriggerrer struct {
	url string
}

// Trigger implemented Triggerrer and actually performs the encapsulated
// behavior.
func (handler *SqsTriggerrer) Trigger(message string) (err error) {
	return errors.New("Not yet implemented")
}

// NewSqsTriggerrer initializes a SqsTriggerrer. It may no longer be
// strictly necessary.
func NewSqsTriggerrer(url string) *SqsTriggerrer {
	var handler SqsTriggerrer
	handler.url = url

	return &handler
}
