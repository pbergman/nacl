package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

type PrivateKey struct {
	key *rsa.PrivateKey
}

func (p *PrivateKey) Sign(payload []byte) (string, error) {

	hash := sha512.New()
	hash.Write(payload)

	signature, err := rsa.SignPKCS1v15(rand.Reader, p.key, crypto.SHA512, hash.Sum(nil))

	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func (p *PrivateKey) UnmarshalText(text []byte) error {
	block, _ := pem.Decode(text)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)

	if err != nil {
		return err
	}

	pri, ok := key.(*rsa.PrivateKey)

	if !ok {
		return fmt.Errorf("invalid key type, expected rsa private key got '%T'", key)
	}

	p.key = pri

	return nil
}
