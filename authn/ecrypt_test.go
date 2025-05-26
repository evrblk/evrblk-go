package authn

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestP256(t *testing.T) {
	require := require.New(t)

	privateKey, publicKey, err := GenerateP256KeyPair()
	require.NoError(err)

	data := []byte("This is a sample text for testing.")
	signature, err := SignP256(data, privateKey)
	require.NoError(err)

	//fmt.Println("P256 signature:", signature)

	err = VerifyP256(data, signature, publicKey)
	require.NoError(err)
}

func BenchmarkSignP256(b *testing.B) {
	// Generate key pairs
	privateKey, _, _ := GenerateP256KeyPair()

	// Create sample data
	data := []byte("This is a sample text for testing.")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Sign the data
		_, err := SignP256(data, privateKey)
		if err != nil {
			b.Fatalf("Signing error: %s", err)
		}
	}
}

func BenchmarkVerifyP256(b *testing.B) {
	// Generate key pairs
	privateKey, publicKey, _ := GenerateP256KeyPair()

	// Create sample data
	data := []byte("This is a sample text for testing.")

	// Sign the data
	signature, err := SignP256(data, privateKey)
	if err != nil {
		b.Fatalf("Signing error: %s", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Verify the signature
		err = VerifyP256(data, signature, publicKey)
		if err != nil {
			b.Fatalf("Verification error: %s", err)
		}
	}
}
