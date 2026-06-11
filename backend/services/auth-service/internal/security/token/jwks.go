package token

import (
	"crypto/rsa"
	"encoding/base64"
	"math/big"
)

type JWKSet struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	PublicUse string `json:"use"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

func BuildJWKS(keys *KeySet) JWKSet {
	if keys == nil {
		return JWKSet{Keys: []JWK{}}
	}

	publicKeys := keys.PublicKeys()
	jwks := JWKSet{Keys: make([]JWK, 0, len(publicKeys))}
	for keyID, publicKey := range publicKeys {
		if publicKey == nil {
			continue
		}
		jwks.Keys = append(jwks.Keys, rsaPublicJWK(keyID, publicKey))
	}
	return jwks
}

func rsaPublicJWK(keyID string, publicKey *rsa.PublicKey) JWK {
	return JWK{
		KeyType:   "RSA",
		KeyID:     keyID,
		PublicUse: "sig",
		Algorithm: SigningAlgorithmRS256,
		Modulus:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
		Exponent:  base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
	}
}
