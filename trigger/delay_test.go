package trigger

import (
	"github.com/stretchr/testify/assert"
	//"github.com/stuphlabs/pullcord"
	"testing"
	"time"
)

func TestDelayTriggerSingleDelay(t *testing.T) {
	testString := "testing"
	cth := &counterTriggerHandler{testString, 0}

	var dt DelayTrigger
	dt.Trigger = cth
	dt.Delay = time.Second

	err := dt.TriggerString(testString)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), cth.count)

	time.Sleep(2*time.Second)

	assert.Equal(t, uint(1), cth.count)
}

func TestDelayTriggerDoubleDelay(t *testing.T) {
	testString := "testing"
	cth := &counterTriggerHandler{testString, 0}

	var dt DelayTrigger
	dt.Trigger = cth
	dt.Delay = 3 * time.Second

	err := dt.TriggerString(testString)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), cth.count)

	time.Sleep(2 * time.Second)
	assert.Equal(t, uint(0), cth.count)
	err = dt.TriggerString(testString)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), cth.count)

	time.Sleep(2 * time.Second)
	// the trigger would have definitely fired by now if the second delay
	// hadn't occurred when it did
	assert.Equal(t, uint(0), cth.count)

	time.Sleep(2*time.Second)
	assert.Equal(t, uint(1), cth.count)
}

func TestDelayTriggerErrorMasking(t *testing.T) {
	testString := "testing"
	cth := &counterTriggerHandler{testString, 0}

	var dt DelayTrigger
	dt.Trigger = cth
	dt.Delay = time.Second

	err := dt.TriggerString("error")
	assert.NoError(t, err)
	assert.Equal(t, uint(0), cth.count)

	time.Sleep(2*time.Second)

	assert.Equal(t, uint(0), cth.count)
}

func TestDelayTriggerReplaceError(t *testing.T) {
	testString := "testing"
	cth := &counterTriggerHandler{testString, 0}

	var dt DelayTrigger
	dt.Trigger = cth
	dt.Delay = 3 * time.Second

	err := dt.TriggerString("error")
	assert.NoError(t, err)
	assert.Equal(t, uint(0), cth.count)

	time.Sleep(2 * time.Second)
	assert.Equal(t, uint(0), cth.count)
	err = dt.TriggerString(testString)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), cth.count)

	time.Sleep(2 * time.Second)
	// the trigger would have definitely fired by now if the second delay
	// hadn't occurred when it did
	assert.Equal(t, uint(0), cth.count)

	time.Sleep(2*time.Second)
	assert.Equal(t, uint(1), cth.count)
}

func TestDelayTriggerIntroduceError(t *testing.T) {
	testString := "testing"
	cth := &counterTriggerHandler{testString, 0}

	var dt DelayTrigger
	dt.Trigger = cth
	dt.Delay = 3 * time.Second

	err := dt.TriggerString(testString)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), cth.count)

	time.Sleep(2 * time.Second)
	assert.Equal(t, uint(0), cth.count)
	err = dt.TriggerString("error")
	assert.NoError(t, err)
	assert.Equal(t, uint(0), cth.count)

	time.Sleep(2 * time.Second)
	// the trigger would have definitely fired by now if the second delay
	// hadn't occurred when it did
	assert.Equal(t, uint(0), cth.count)

	time.Sleep(2*time.Second)
	assert.Equal(t, uint(0), cth.count)
}

