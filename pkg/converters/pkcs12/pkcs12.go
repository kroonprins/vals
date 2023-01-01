package pkcs12

import (
	"github.com/kroonprins/vals/pkg/api"

	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

type provider struct {
	password string
}
type toCert struct {
	provider
}
type toKey struct {
	provider
}

func NewToCert(cfg api.StaticConfig) *toCert {
	p := &toCert{
		provider: new(cfg),
	}
	return p
}
func NewToKey(cfg api.StaticConfig) *toKey {
	p := &toKey{
		provider: new(cfg),
	}
	return p
}

func new(cfg api.StaticConfig) provider {
	return provider{
		password: cfg.String("password"),
	}
}

func (p *toCert) Convert(value string) (string, error) {
	_, certs, err := toPEM(value, p.password)
	if err != nil {
		return "", fmt.Errorf("failure extracting certificate from pkcs12: %v", err)
	}
	return string(certs), nil
}

func (p *toKey) Convert(value string) (string, error) {
	key, _, err := toPEM(value, p.password)
	if err != nil {
		return "", fmt.Errorf("failure extracting key from pkcs12: %v", err)
	}
	return string(key), nil
}

func toPEM(fileContent string, password string) ([]byte, []byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(fileContent)
	if err != nil {
		decoded = []byte(fileContent)
	}

	privateKey, certificate, chain, err := pkcs12.DecodeChain(decoded, password)
	if err != nil {
		return nil, nil, fmt.Errorf("failure decoding pkcs12: %v", err)
	}

	x509Certs := []*x509.Certificate{certificate}
	x509Certs = append(x509Certs, chain...)

	certs := []byte{}
	for _, certPem := range x509Certs {
		certs = append(certs, pem.EncodeToMemory(
			&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: certPem.Raw,
			},
		)...)
	}

	var privateKeyBytes []byte
	switch typed_privateKey := privateKey.(type) {
	case *rsa.PrivateKey:
		privateKeyBytes = x509.MarshalPKCS1PrivateKey(typed_privateKey)
	case *ecdsa.PrivateKey:
		privateKeyBytes, err = x509.MarshalECPrivateKey(typed_privateKey)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to marshal ec private key: %v", err)
		}
	case string:
		privateKeyBytes, err = x509.MarshalPKCS8PrivateKey(typed_privateKey)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to marshal pkcs8 private key: %v", err)
		}
	default:
		return nil, nil, fmt.Errorf("unhandled private key type: %T", privateKey)
	}

	key := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)

	return key, certs, nil
}
