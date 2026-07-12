package configuration

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

type RSAPrivateKey rsa.PrivateKey

func (k *RSAPrivateKey) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	pemBytes, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return fmt.Errorf("base64 decode: %w", err)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return errors.New("no PEM block found (is it really PEM?)")
	}

	key, err := parseRSAKey(block)
	if err != nil {
		return err
	}
	*k = RSAPrivateKey(*key)
	return nil
}

func parseRSAKey(block *pem.Block) (*rsa.PrivateKey, error) {
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)

	case "PRIVATE KEY":
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("expected RSA key, got %T", parsed)
		}
		return rsaKey, nil

	default:
		return nil, fmt.Errorf("unsupported PEM type %q", block.Type)
	}
}
