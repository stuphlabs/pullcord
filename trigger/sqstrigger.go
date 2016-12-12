package trigger

import (
	"errors"
	// "github.com/stuphlabs/pullcord"
)

type SqsTriggerHandler struct {
	url string
}

func (handler *SqsTriggerHandler) Trigger(message string) (err error) {
	return errors.New("Not yet implemented")
}

func NewSqsTriggerHandler(url string) *SqsTriggerHandler {
	var handler SqsTriggerHandler
	handler.url = url

	return &handler
}
