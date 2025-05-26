package authn

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"github.com/go-errors/errors"
	"math/big"
)

type ECDSASignature struct {
	R, S *big.Int
}

// GenerateP256KeyPair generates a new secp256r1 key pair
func GenerateP256KeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, &privateKey.PublicKey, nil
}

// SignP256 signs the given data with a private key
func SignP256(data []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	// Compute the SHA-256 hash of the data
	hash := sha256.Sum256(data)

	// Sign the hash using the private key
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// VerifyP256 checks whether the signature is valid for the given data and public key
func VerifyP256(data []byte, signature []byte, publicKey *ecdsa.PublicKey) error {
	hash := sha256.Sum256(data)

	sig := ECDSASignature{}
	_, err := asn1.Unmarshal(signature, &sig)
	if err != nil {
		return err
	}

	valid := ecdsa.Verify(publicKey, hash[:], sig.R, sig.S)
	if valid {
		return nil
	} else {
		return errors.New("invalid signature")
	}
}
