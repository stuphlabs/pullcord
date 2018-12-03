package trigger

import (
	"errors"
	// "github.com/stuphlabs/pullcord"
)

// SqsTriggerHandler is a work-in-progress that is intended to eventually send
// SQS messages as the result of a triggering event.
type SqsTriggerHandler struct {
	url string
}

func (handler *SqsTriggerHandler) Trigger(message string) (err error) {
	return errors.New("Not yet implemented")
}

// NewSqsTriggerHandler initializes a SqsTriggerHandler. It may no longer be
// strictly necessary.
func NewSqsTriggerHandler(url string) *SqsTriggerHandler {
	var handler SqsTriggerHandler
	handler.url = url

	return &handler
}
