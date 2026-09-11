package authn

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeVTMessage struct{}

func (fakeVTMessage) MarshalVT() ([]byte, error) {
	return []byte("payload"), nil
}

func TestVerifyAlfaSignature_MalformedPublicPemReturnsError(t *testing.T) {
	now := time.Now()

	err := VerifyAlfaSignature("c2ln", now.Unix(), now, "not a pem", fakeVTMessage{}, "Moab", "CreateQueue")

	require.Error(t, err)
	assert.NotPanics(t, func() {
		_ = VerifyAlfaSignature("c2ln", now.Unix(), now, "not a pem", fakeVTMessage{}, "Moab", "CreateQueue")
	})
}

func TestSignAlfa_MalformedPrivatePemReturnsError(t *testing.T) {
	now := time.Now()

	_, err := SignAlfa(now.Unix(), "not a pem", fakeVTMessage{}, "Moab", "CreateQueue")

	require.Error(t, err)
	assert.NotPanics(t, func() {
		_, _ = SignAlfa(now.Unix(), "not a pem", fakeVTMessage{}, "Moab", "CreateQueue")
	})
}
