package net

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/stuphlabs/pullcord/config"
)

func init() {
	e := config.RegisterResourceType(
		"testcertgetter",
		func() json.Unmarshaler {
			return new(TestCertificateGetter)
		},
	)

	if e != nil {
		panic(e)
	}
}

type TestCertificateGetter struct{
	Cert *tls.Certificate
}

func (t *TestCertificateGetter) UnmarshalJSON([]byte) error {
	return nil
}

func (t *TestCertificateGetter) GetCertificate(
	*tls.ClientHelloInfo,
) (*tls.Certificate, error) {
	if t.Cert == nil {
		return nil, errors.New(
			"No certificate was added to this" +
				" TestCertificateGetter",
		)
	} else {
		return t.Cert, nil
	}
}

type EavesdroppedConn struct {
	Target net.Conn
	Inbound io.Writer
	Outbound io.Writer
}

func (ec EavesdroppedConn) Read(b []byte) (int, error) {
	n, er := ec.Target.Read(b)
	if n <= 0 {
		return n, er
	}

	oer := er

	n, er = ec.Inbound.Write(b[:n])

	if oer != nil {
		er = oer
	}

	return n, er
}

func (ec EavesdroppedConn) Write(b []byte) (int, error) {
	n, er := ec.Target.Write(b)
	if n <= 0 {
		return n, er
	}

	oer := er

	n, er = ec.Outbound.Write(b[:n])

	if oer != nil {
		er = oer
	}

	return n, er
}

func (ec EavesdroppedConn) Close() error {
	er := ec.Target.Close()

	if er == nil {
		if oc, ok := ec.Outbound.(io.Closer); ok {
			er = oc.Close()
		}
	}

	if er == nil {
		if ic, ok := ec.Inbound.(io.Closer); ok {
			er = ic.Close()
		}
	}

	return er
}

func (ec EavesdroppedConn) LocalAddr() net.Addr {
	return ec.Target.LocalAddr()
}

func (ec EavesdroppedConn) RemoteAddr() net.Addr {
	return ec.Target.RemoteAddr()
}

func (ec EavesdroppedConn) SetDeadline(t time.Time) error {
	return ec.Target.SetDeadline(t)
}

func (ec EavesdroppedConn) SetReadDeadline(t time.Time) error {
	return ec.Target.SetReadDeadline(t)
}

func (ec EavesdroppedConn) SetWriteDeadline(t time.Time) error {
	return ec.Target.SetWriteDeadline(t)
}

type BufferedEavesdropListener struct {
	Target net.Listener
	Inbound bytes.Buffer
	Outbound bytes.Buffer
}

func (bel *BufferedEavesdropListener) Accept() (net.Conn, error) {
	c, e := bel.Target.Accept()

	ec := EavesdroppedConn{
		Target: c,
		Inbound: &bel.Inbound,
		Outbound: &bel.Outbound,
	}

	return ec, e
}

func (bel *BufferedEavesdropListener) Close() error {
	return bel.Target.Close()
}

func (bel *BufferedEavesdropListener) Addr() net.Addr {
	return bel.Target.Addr()
}


func GenSelfSignedLocalhostCertificate(
	validFor time.Duration,
) (*tls.Certificate, *x509.Certificate, error) {
	now := time.Now()

	tpl := &x509.Certificate{
		DNSNames: []string{"localhost", "localhost.localdomain"},
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
		},
		IsCA: true,
		NotBefore: now,
		NotAfter: now.Add(validFor),
		PublicKeyAlgorithm: x509.RSA,
		SerialNumber: big.NewInt(1),
		SignatureAlgorithm: x509.SHA512WithRSA,
	}

	privKey, e := rsa.GenerateKey(rand.Reader, 1024)
	if e != nil {
		return nil, nil, errors.Wrap(
			e,
			"Unable to gen RSA key for self-signed cert",
		)
	}

	derCert, e := x509.CreateCertificate(
		rand.Reader,
		tpl,
		tpl,
		&privKey.PublicKey,
		privKey,
	)
	if e != nil {
		return nil, nil, errors.Wrap(
			e,
			"Unable to get raw form of self-signed cert",
		)
	}

	pemCert := pem.EncodeToMemory(
		&pem.Block{
			Type: "CERTIFICATE",
			Bytes: derCert,
		},
	)

	pemKey := pem.EncodeToMemory(
		&pem.Block{
			Type: "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privKey),
		},
	)

	x509Cert, e := x509.ParseCertificate(derCert)
	if e != nil {
		return nil, nil, errors.Wrap(
			e,
			"Unable to get x509 form of self-signed cert",
		)
	}

	tlsCert, e := tls.X509KeyPair(pemCert, pemKey)
	if e != nil {
		return nil, nil, errors.Wrap(
			e,
			"Unable to get tls form of self-signed cert",
		)
	}

	return &tlsCert, x509Cert, nil
}
