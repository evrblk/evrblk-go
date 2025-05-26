package authn

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"
	"time"
)

func VerifyAlfaSignature(signatureBase64 string, timestamp int64, now time.Time, publicPem string, request proto.Message) error {
	// Check timestamp for replays
	ts := time.Unix(timestamp, 0)
	if now.Add(time.Minute*5).Before(ts) || now.Add(time.Minute*-5).After(ts) {
		return errors.New("invalid timestamp")
	}

	// Decode signature from Base64
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return err
	}

	// Serialize timestamp and request body
	requestBytes, err := serializeRequest(request)
	if err != nil {
		return err
	}
	timestampBytes := serializeTimestamp(timestamp)
	data := append(timestampBytes, requestBytes...)

	// Deserialize public PEM string
	block, _ := pem.Decode([]byte(publicPem))

	// Deserialize ECDSA public key
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	ecdsaPublicKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("not an ECDSA public key")
	}

	// Verify timestamped request
	return VerifyP256(data, signature, ecdsaPublicKey)
}

func GenerateAlfaKeys() (privatePem string, publicPem string, err error) {
	privateKey, publicKey, err := GenerateP256KeyPair()
	if err != nil {
		return
	}

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return
	}
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return
	}

	privateBlock := &pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyBytes}
	privatePem = string(pem.EncodeToMemory(privateBlock))

	publicBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyBytes}
	publicPem = string(pem.EncodeToMemory(publicBlock))

	return
}

func SignAlfa(timestamp int64, privatePem string, request proto.Message) (string, error) {
	// Serialize timestamp and request body
	requestBytes, err := serializeRequest(request)
	if err != nil {
		return "", err
	}
	timestampBytes := serializeTimestamp(timestamp)
	data := append(timestampBytes, requestBytes...)

	// Deserialize private PEM string
	block, _ := pem.Decode([]byte(privatePem))

	// Deserialize ECDSA private key
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	// Sign
	signature, err := SignP256(data, privateKey)
	if err != nil {
		return "", err
	}

	// Return Base64 of signature
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)

	return signatureBase64, nil
}

// serializeRequest serializes protobuf body in a deterministic way
func serializeRequest(request proto.Message) ([]byte, error) {
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(request)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// serializeTimestamp serializes timestamp as 8 bytes big endian integer
func serializeTimestamp(timestamp int64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(timestamp))
	return data
}
