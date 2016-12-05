package trigger

import (
	"errors"
	"github.com/stretchr/testify/assert"
	//"github.com/stuphlabs/pullcord"
	"testing"
)

type counterTriggerHandler struct {
	name string
	count uint
}

func (th *counterTriggerHandler) TriggerString(name string) error {
	if name == th.name {
		th.count += 1
		return nil
	} else {
		return errors.New("bad error string")
	}
}

func TestCompoundTriggerNoErrors(t *testing.T) {
	testString := "testing"

	th1 := &counterTriggerHandler{testString, 0}
	th2 := &counterTriggerHandler{testString, 0}

	ct := CompoundTrigger{[]TriggerHandler{th1, th2}}

	err := ct.TriggerString(testString)
	assert.NoError(t, err)

	err = ct.TriggerString(testString)
	assert.NoError(t, err)

	assert.Equal(t, uint(2), th1.count)
	assert.Equal(t, uint(2), th2.count)
}

func TestCompoundTriggerAllErrors(t *testing.T) {
	testString := "testing"

	th1 := &counterTriggerHandler{testString, 0}
	th2 := &counterTriggerHandler{testString, 0}

	ct := CompoundTrigger{[]TriggerHandler{th1, th2}}

	err := ct.TriggerString("error")
	assert.Error(t, err)

	assert.Equal(t, uint(0), th1.count)
	assert.Equal(t, uint(0), th2.count)
}

func TestCompoundTriggerSomeErrors(t *testing.T) {
	testString := "testing"

	th1 := &counterTriggerHandler{testString, 0}
	th2 := &counterTriggerHandler{"error", 0}

	ct := CompoundTrigger{[]TriggerHandler{th1, th2}}

	err := ct.TriggerString(testString)
	assert.Error(t, err)

	assert.Equal(t, uint(1), th1.count)
	assert.Equal(t, uint(0), th2.count)
}

