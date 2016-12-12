package trigger

import (
	"errors"
	"github.com/stretchr/testify/assert"
	//"github.com/stuphlabs/pullcord"
	"testing"
)

type counterTriggerHandler struct {
	count int
}

func (th *counterTriggerHandler) Trigger() error {
	if th.count >= 0 {
		th.count += 1
		return nil
	} else {
		return errors.New("this trigger always errors")
	}
}

func TestCompoundTriggerNoErrors(t *testing.T) {
	th1 := &counterTriggerHandler{}
	th2 := &counterTriggerHandler{}

	ct := CompoundTrigger{[]TriggerHandler{th1, th2}}

	err := ct.Trigger()
	assert.NoError(t, err)

	err = ct.Trigger()
	assert.NoError(t, err)

	assert.Equal(t, 2, th1.count)
	assert.Equal(t, 2, th2.count)
}

func TestCompoundTriggerAllErrors(t *testing.T) {
	th1 := &counterTriggerHandler{-1}
	th2 := &counterTriggerHandler{-1}

	ct := CompoundTrigger{[]TriggerHandler{th1, th2}}

	err := ct.Trigger()
	assert.Error(t, err)

	assert.Equal(t, -1, th1.count)
	assert.Equal(t, -1, th2.count)
}

func TestCompoundTriggerSomeErrors(t *testing.T) {
	th1 := &counterTriggerHandler{}
	th2 := &counterTriggerHandler{-1}

	ct := CompoundTrigger{[]TriggerHandler{th1, th2}}

	err := ct.Trigger()
	assert.Error(t, err)

	assert.Equal(t, 1, th1.count)
	assert.Equal(t, -1, th2.count)
}

