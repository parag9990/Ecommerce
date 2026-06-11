package token

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

const SigningAlgorithmRS256 = "RS256"

type KeySet struct {
	keyID      string
	privateKey *rsa.PrivateKey
	publicKeys map[string]*rsa.PublicKey
}

func LoadRSAKeySet(keyID string, privateKeyPath string, publicKeyPath string) (*KeySet, error) {
	if keyID == "" {
		return nil, errors.New("jwt key id is required")
	}
	if privateKeyPath == "" {
		return nil, errors.New("jwt private key path is required")
	}

	privatePEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read jwt private key: %w", err)
	}
	privateKey, err := ParseRSAPrivateKeyPEM(privatePEM)
	if err != nil {
		return nil, err
	}

	publicKey := &privateKey.PublicKey
	if publicKeyPath != "" {
		publicPEM, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("read jwt public key: %w", err)
		}
		publicKey, err = ParseRSAPublicKeyPEM(publicPEM)
		if err != nil {
			return nil, err
		}
	}

	return NewKeySet(keyID, privateKey, publicKey), nil
}

func NewKeySet(keyID string, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *KeySet {
	keys := map[string]*rsa.PublicKey{}
	if publicKey != nil {
		keys[keyID] = publicKey
	}
	return &KeySet{keyID: keyID, privateKey: privateKey, publicKeys: keys}
}

func (k *KeySet) KeyID() string {
	return k.keyID
}

func (k *KeySet) PrivateKey() *rsa.PrivateKey {
	return k.privateKey
}

func (k *KeySet) PublicKeyByID(keyID string) (*rsa.PublicKey, error) {
	if k == nil {
		return nil, errors.New("key set is required")
	}
	publicKey, ok := k.publicKeys[keyID]
	if !ok || publicKey == nil {
		return nil, fmt.Errorf("unknown jwt key id %q", keyID)
	}
	return publicKey, nil
}

func (k *KeySet) PublicKeys() map[string]*rsa.PublicKey {
	keys := make(map[string]*rsa.PublicKey, len(k.publicKeys))
	for keyID, publicKey := range k.publicKeys {
		keys[keyID] = publicKey
	}
	return keys
}

func ParseRSAPrivateKeyPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid rsa private key pem")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse rsa private key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not rsa")
	}
	return key, nil
}

func ParseRSAPublicKeyPEM(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid rsa public key pem")
	}

	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("public key is not rsa")
		}
		return rsaKey, nil
	}

	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse rsa public key: %w", err)
	}
	return key, nil
}
