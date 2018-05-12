package net

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	configutil "github.com/stuphlabs/pullcord/config/util"
)

func TestBasicTlsListenerType(t *testing.T) {
	b := new(BasicTlsListener)

	assert.Implements(
		t,
		(*net.Listener)(nil),
		b,
		"A BasicTlsListener should be a net.Listener, but this"+
			" BasicTlsListener has the wrong type.",
	)
}

func TestBasicTlsListenerConfig(t *testing.T) {
	test := configutil.ConfigTest{
		ResourceType: "basictlslistener",
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
					"listener": null,
					"certgetter": null
				}`,
				Explanation: "bad protocol",
			},
			{
				Data: `{
					"listener": {
						"type": "testcertgetter",
						"data": null
					},
					"certgetter": {
						"type": "testcertgetter",
						"data": null
					}
				}`,
				Explanation: "bad listener",
			},
			{
				Data: `{
					"listener": {
						"type": "testlistener",
						"data": null
					},
					"certgetter": {
						"type": "testlistener",
						"data": null
					}
				}`,
				Explanation: "bad certgetter",
			},
		},
		Good: []configutil.ConfigTestData{
			{
				Data: `{
					"listener": {
						"type": "testlistener",
						"data": null
					},
					"certgetter": {
						"type": "testcertgetter",
						"data": null
					}
				}`,
				Explanation: "good config",
			},
		},
	}

	test.Run(t)
}

func TestBasicTlsListenerBehavior(t *testing.T) {
	nl, e := net.Listen("tcp", "127.0.0.1:0")
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

	tlsCert, x509Cert, e := GenSelfSignedLocalhostCertificate(
		3 * time.Minute,
	)
	require.NoError(t, e, "Valid certificates are needed for testing.")

	bel := &BufferedEavesdropListener{
		Target: nl,
	}

	l := &BasicTlsListener{
		Listener:   bel,
		CertGetter: &TestCertificateGetter{Cert: tlsCert},
	}

	if l == nil {
		return
	}

	defer func() {
		e := l.Close()
		assert.NoError(
			t,
			e,
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
		l *BasicTlsListener,
		bel *BufferedEavesdropListener,
		expected string,
		done chan<- interface{},
	) {
		defer func(done chan<- interface{}) {
			close(done)
		}(done)

		c, e := l.Accept()
		assert.NoError(
			t,
			e,
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

		endToEndBuffer := new(bytes.Buffer)

		_, e = endToEndBuffer.ReadFrom(c)
		assert.NoError(
			t,
			e,
			"Attempting to read from an accepted connection that"+
				" came from a valid basic listener should not"+
				" produce an error, but an error was produced.",
		)

		actualEndToEndMsg := endToEndBuffer.String()
		assert.Equal(
			t,
			expected,
			actualEndToEndMsg,
			"Attempting to transmit data through a basic"+
				" listener should not result in alterred data, but"+
				" the actual data received was not identical to the"+
				" data expected to have been sent.",
		)

		actualTransportMsg := bel.Inbound.String()
		assert.NotEqual(
			t,
			expected,
			actualTransportMsg,
			"Transmitting data through a basic TLS listener"+
				" should not result in the original message"+
				" being sent in the clear of the raw"+
				" transport layer.",
		)
	}(t, l, bel, expected, done)

	certPool := x509.NewCertPool()
	certPool.AddCert(x509Cert)

	c, e := tls.Dial(
		addr.Network(),
		addr.String(),
		&tls.Config{
			RootCAs: certPool,
		},
	)
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
		e := c.Close()
		assert.NoError(
			t,
			e,
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
