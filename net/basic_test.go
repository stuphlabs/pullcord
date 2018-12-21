package net

import (
	"bytes"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
)

func TestBasicListenerType(t *testing.T) {
	b := new(BasicListener)

	assert.Implements(
		t,
		(*net.Listener)(nil),
		b,
		"A BasicListener should be a net.Listener, but this"+
			" BasicListener has the wrong type.",
	)
}

func TestBasicListenerConfig(t *testing.T) {
	test := configutil.ConfigTest{
		ResourceType: "basiclistener",
		ListenerTest: true,
		SyntacticallyBad: []configutil.ConfigTestData{
			{
				Data:        ``,
				Explanation: "empty config",
			},
			{
				Data:        `42`,
				Explanation: "numeric config",
			},
			{
				Data:        `"test"`,
				Explanation: "string config",
			},
			{
				Data: `{
				}`,
				Explanation: "empty object",
			},
			{
				Data: `{
					"proto": "droid",
					"laddr": ":0"
				}`,
				Explanation: "bad protocol",
			},
			{
				Data: `{
					"proto": "tcp",
					"laddr": "1234 Main St. Podunk, OK 73000"
				}`,
				Explanation: "bad address",
			},
		},
		Good: []configutil.ConfigTestData{
			{
				Data: `{
					"proto": "tcp",
					"laddr": ":0"
				}`,
				Explanation: "good config",
			},
		},
	}

	test.Run(t)
}

func TestBasicListenerBehavior(t *testing.T) {
	var nl net.Listener
	var e error
	nl, e = net.Listen("tcp", ":0")
	assert.NoError(
		t,
		e,
		"Attempting to create a valid basic listener should not"+
			" produce an error, but an error was produced.",
	)
	assert.NotNil(
		t,
		nl,
		"Attempting to create a valid basic listener should produce"+
			" a non-nil listener, but a nil listener was produced.",
	)

	l := &BasicListener{nl}

	if l == nil {
		return
	}

	defer func() {
		err := l.Close()
		assert.NoError(
			t,
			err,
			"Attempting to close a basic listener should not"+
				" produce an error, but an error was produced.",
		)
	}()

	addr := l.Addr()
	assert.NotNil(
		t,
		addr,
		"Attempting to get the address from a basic listener should"+
			" return a non-nil address, but a nil address was returned.",
	)

	if addr == nil {
		return
	}

	expected := "testing"

	done := make(chan interface{})
	defer func(done <-chan interface{}) {
		<-done
	}(done)

	go func(
		t *testing.T,
		l *BasicListener,
		expected string,
		done chan<- interface{},
	) {
		defer func(done chan<- interface{}) {
			close(done)
		}(done)

		c, err := l.Accept()
		assert.NoError(
			t,
			err,
			"Attempting to accept a connection from a valid"+
				" basic listener should not produce an error, but an"+
				" error was produced.",
		)
		assert.NotNil(
			t,
			c,
			"Attempting to accept a connection from a valid"+
				" basic listener should produce a non-nil"+
				" connection, but a nil connection was produced.",
		)
		if c == nil {
			return
		}

		b := new(bytes.Buffer)

		_, err = b.ReadFrom(c)
		assert.NoError(
			t,
			err,
			"Attempting to read from an accepted connection that"+
				" came from a valid basic listener should not"+
				" produce an error, but an error was produced.",
		)

		actual := b.String()
		assert.Equal(
			t,
			expected,
			actual,
			"Attempting to transmit data through a basic"+
				" listener should not result in alterred data, but"+
				" the actual data received was not identical to the"+
				" data expected to have been sent.",
		)
	}(t, l, expected, done)

	var c net.Conn
	c, e = net.Dial(addr.Network(), addr.String())
	assert.NoError(
		t,
		e,
		"Attempting to connect as a client to a basic listener"+
			" should not produce an error, but an error was produced.",
	)
	assert.NotNil(
		t,
		c,
		"Attempting to connect as a client to a basic listener"+
			" should produce a non-nil connection, but a nil connection"+
			" was produced.",
	)

	if c == nil {
		return
	}

	defer func(t *testing.T, c net.Conn) {
		err := c.Close()
		assert.NoError(
			t,
			err,
			"Attempting to close a client connection to a basic"+
				" listener should not produce an error, but an error"+
				" was produced.",
		)
	}(t, c)

	_, e = c.Write([]byte(expected))
	assert.NoError(
		t,
		e,
		"Attempting to write to a client connection of a basic"+
			" listener should not produce an error, but an error was"+
			" produced.",
	)
}
