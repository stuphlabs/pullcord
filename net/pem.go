package net

import (
	"golang.org/x/crypto/acme/autocert"
)

type PemConfig struct {
	Cert []byte
	Key []byte
}

func (p *PemConfig) GetCertificate() (Certificate, error) {
	return tls.X509KeyPair(a.Cert, a.Key)
}
