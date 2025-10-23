package authn

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"
)

func VerifyBravoSignature(signatureHex string, timestamp int64, now time.Time, hashedSecret []byte, request proto.Message, service string, method string) error {
	// Check timestamp for replays
	ts := time.Unix(timestamp, 0)
	if now.Add(time.Minute*5).Before(ts) || now.Add(time.Minute*-5).After(ts) {
		return errors.New("invalid timestamp")
	}

	// Serialize timestamp and request body
	requestBytes, err := serializeRequest(request)
	if err != nil {
		return err
	}
	timestampBytes := serializeTimestamp(timestamp)

	data := append(timestampBytes, []byte(fmt.Sprintf("%s.%s", service, method))...)
	data = append(data, requestBytes...)

	// Decode signature from HEX
	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return err
	}

	// Verify timestamped request
	if verifyHMAC(hashedSecret, data, signature) {
		return nil
	} else {
		return errors.New("signature mismatch")
	}
}

func GenerateBravoSecret() string {
	// Generate random bytes
	buf := make([]byte, 512)
	_, err := rand.Read(buf)
	if err != nil {
		log.Fatalf("error while generating random string: %s", err)
	}

	// Return Base64 of those bytes
	return base64.StdEncoding.EncodeToString(buf)
}

func SignBravo(timestamp int64, secretBase64 string, request proto.Message, service string, method string) (string, error) {
	// Serialize timestamp and request body
	requestBytes, err := serializeRequest(request)
	if err != nil {
		return "", err
	}
	timestampBytes := serializeTimestamp(timestamp)

	data := append(timestampBytes, []byte(fmt.Sprintf("%s.%s", service, method))...)
	data = append(data, requestBytes...)

	// Hash secret with a date
	date := GetDateOfTimestamp(timestamp)
	hashedSecret, err := HashBravoSecretWithDate(secretBase64, date)
	if err != nil {
		return "", err
	}

	// Sign
	signature, err := generateHMAC(hashedSecret, data)
	if err != nil {
		return "", err
	}

	// Return HEX string of signature
	return hex.EncodeToString(signature), nil
}

func GetDateOfTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2006-01-02")
}

func HashBravoSecretWithDate(secretBase64 string, date string) ([]byte, error) {
	secret, err := base64.StdEncoding.DecodeString(secretBase64)
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	_, err = h.Write([]byte(date))
	if err != nil {
		return nil, err
	}
	_, err = h.Write(secret)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func generateHMAC(key []byte, data []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(data)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func verifyHMAC(key []byte, data []byte, signature []byte) bool {
	expectedSignature, err := generateHMAC(key, data)
	if err != nil {
		return false
	}
	return hmac.Equal(signature, expectedSignature)
}
