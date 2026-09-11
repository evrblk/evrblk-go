package evrblk

import (
	"testing"

	"github.com/evrblk/evrblk-go/authn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validAlfaKeyPair(t *testing.T) (privatePem string, publicPem string) {
	t.Helper()
	privatePem, publicPem, err := authn.GenerateAlfaKeys()
	require.NoError(t, err)
	return privatePem, publicPem
}

func validBravoSecret(t *testing.T) string {
	t.Helper()
	secret, err := authn.GenerateBravoSecret()
	require.NoError(t, err)
	return secret
}

func TestNewAlfaRequestSigner(t *testing.T) {
	privatePem, _ := validAlfaKeyPair(t)

	t.Run("valid key id and pem succeeds", func(t *testing.T) {
		signer, err := NewAlfaRequestSigner("key_alfa_abc123", privatePem)
		require.NoError(t, err)
		assert.NotNil(t, signer)
	})

	t.Run("rejects a non-alfa key id", func(t *testing.T) {
		_, err := NewAlfaRequestSigner("key_bravo_abc123", privatePem)
		assert.Error(t, err)
	})

	t.Run("rejects a malformed private pem", func(t *testing.T) {
		_, err := NewAlfaRequestSigner("key_alfa_abc123", "not a pem")
		assert.Error(t, err)
	})
}

func TestNewBravoRequestSigner(t *testing.T) {
	secret := validBravoSecret(t)

	t.Run("valid key id and secret succeeds", func(t *testing.T) {
		signer, err := NewBravoRequestSigner("key_bravo_abc123", secret)
		require.NoError(t, err)
		assert.NotNil(t, signer)
	})

	t.Run("rejects a non-bravo key id", func(t *testing.T) {
		_, err := NewBravoRequestSigner("key_alfa_abc123", secret)
		assert.Error(t, err)
	})

	t.Run("rejects a malformed secret", func(t *testing.T) {
		_, err := NewBravoRequestSigner("key_bravo_abc123", "not base64!!!")
		assert.Error(t, err)
	})

	t.Run("rejects an empty secret", func(t *testing.T) {
		_, err := NewBravoRequestSigner("key_bravo_abc123", "")
		assert.Error(t, err)
	})
}

func TestNewRequestSigner(t *testing.T) {
	privatePem, _ := validAlfaKeyPair(t)
	secret := validBravoSecret(t)

	t.Run("dispatches alfa key ids to the alfa signer", func(t *testing.T) {
		signer, err := NewRequestSigner("key_alfa_abc123", privatePem)
		require.NoError(t, err)
		assert.IsType(t, &alfaRequestSigner{}, signer)
	})

	t.Run("dispatches bravo key ids to the bravo signer", func(t *testing.T) {
		signer, err := NewRequestSigner("key_bravo_abc123", secret)
		require.NoError(t, err)
		assert.IsType(t, &bravoRequestSigner{}, signer)
	})

	t.Run("rejects charlie key ids, which aren't supported yet", func(t *testing.T) {
		_, err := NewRequestSigner("key_charlie_abc123", secret)
		assert.Error(t, err)
	})

	t.Run("rejects an unrecognized key id", func(t *testing.T) {
		_, err := NewRequestSigner("not-a-key-id", secret)
		assert.Error(t, err)
	})
}
