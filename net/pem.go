package net

import (
	"crypto/tls"
	"encoding/json"

	"github.com/stuphlabs/pullcord/config"
)

func init() {
	e := config.RegisterResourceType(
		"pem",
		func() json.Unmarshaler {
			return new(PemConfig)
		},
	)

	if e != nil {
		panic(e)
	}
}

// PemConfig implements a TlsCertificateGetter using a single PEM encoded key
// and certificate pair.
type PemConfig struct {
	Cert []byte
	Key  []byte
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (p *PemConfig) UnmarshalJSON(d []byte) error {
	var t struct {
		Cert string
		Key  string
	}

	if e := json.Unmarshal(d, &t); e != nil {
		return e
	}

	p.Cert = []byte(t.Cert)
	p.Key = []byte(t.Key)

	return nil
}

// GetCertificate implements TlsCertificateGetter.
func (p *PemConfig) GetCertificate(
	*tls.ClientHelloInfo,
) (*tls.Certificate, error) {
	c, e := tls.X509KeyPair(p.Cert, p.Key)
	return &c, e
}
