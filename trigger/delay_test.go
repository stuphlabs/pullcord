package trigger

import (
	"github.com/stretchr/testify/assert"
	//"github.com/stuphlabs/pullcord"
	"testing"
	"time"
)

func TestDelayTriggerSingleDelay(t *testing.T) {
	cth := &counterTriggerHandler{}

	dt := NewDelayTrigger(cth, time.Second)

	err := dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 0, cth.count)

	time.Sleep(2*time.Second)

	assert.Equal(t, 1, cth.count)
}

func TestDelayTriggerDoubleDelay(t *testing.T) {
	cth := &counterTriggerHandler{}

	dt := NewDelayTrigger(
		cth,
		3 * time.Second,
	)

	err := dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 0, cth.count)

	time.Sleep(2 * time.Second)
	assert.Equal(t, 0, cth.count)
	err = dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 0, cth.count)

	time.Sleep(2 * time.Second)
	// the trigger would have definitely fired by now if the second delay
	// hadn't occurred when it did
	assert.Equal(t, 0, cth.count)

	time.Sleep(2*time.Second)
	assert.Equal(t, 1, cth.count)
}

func TestDelayTriggerErrorMasking(t *testing.T) {
	cth := &counterTriggerHandler{-1}

	dt := NewDelayTrigger(
		cth,
		time.Second,
	)

	err := dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, -1, cth.count)

	time.Sleep(2*time.Second)

	assert.Equal(t, -1, cth.count)
}

func TestDelayTriggerReplaceError(t *testing.T) {
	cth := &counterTriggerHandler{-1}

	dt := NewDelayTrigger(
		cth,
		3 * time.Second,
	)

	err := dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, -1, cth.count)

	time.Sleep(2 * time.Second)
	assert.Equal(t, -1, cth.count)
	err = dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, -1, cth.count)

	time.Sleep(2 * time.Second)
	// the trigger would have definitely fired by now if the second delay
	// hadn't occurred when it did
	assert.Equal(t, -1, cth.count)

	// removes error situation
	cth.count = 0
	assert.Equal(t, 0, cth.count)

	time.Sleep(2*time.Second)
	assert.Equal(t, 1, cth.count)
}

func TestDelayTriggerIntroduceError(t *testing.T) {
	cth := &counterTriggerHandler{}

	dt := NewDelayTrigger(
		cth,
		3 * time.Second,
	)

	err := dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 0, cth.count)

	time.Sleep(2 * time.Second)
	assert.Equal(t, 0, cth.count)
	err = dt.Trigger()
	assert.NoError(t, err)
	assert.Equal(t, 0, cth.count)

	// introduces error situation
	cth.count = -1
	assert.Equal(t, -1, cth.count)

	time.Sleep(2 * time.Second)
	// the trigger would have definitely fired by now if the second delay
	// hadn't occurred when it did
	assert.Equal(t, -1, cth.count)

	time.Sleep(2*time.Second)
	assert.Equal(t, -1, cth.count)
}

