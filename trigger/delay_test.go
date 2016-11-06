package trigger

import (
	"errors"
	"github.com/stretchr/testify/assert"
	//"github.com/stuphlabs/pullcord"
	"testing"
	"time"
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

func TestDelayTrigger(t *testing.T) {
	testString := "testing"
	cth := &counterTriggerHandler{testString, 0}

	var dt DelayTrigger
	dt.Trigger = cth
	dt.Delay = time.Second

	err := dt.TriggerString(testString)
	assert.Error(t, err)
	assert.Equal(t, uint(1), cth.count)

	time.Sleep(time.Second)

	assert.Equal(t, uint(2), cth.count)
}

