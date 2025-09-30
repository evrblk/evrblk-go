package test

import (
	"testing"
	"time"

	"github.com/evrblk/evrblk-go/authn"
	moab "github.com/evrblk/evrblk-go/moab/preview"
	"github.com/labstack/gommon/random"
	"github.com/stretchr/testify/require"
)

// TestAlfaSignAndVerify tests Alfa signing mechanism on a random timestamp, request and secret. A signature should be
// valid within 5 minutes time window.
func TestAlfaSignAndVerify(t *testing.T) {
	require := require.New(t)

	// Get current time
	now := time.Now()
	timestamp := now.Unix()

	// Generate a new Alfa key pair
	privatePem, publicPem, err := authn.GenerateAlfaKeys()
	require.NoError(err)

	// Build a random request with non-trivial values and nested objects
	request := &moab.CreateQueueRequest{
		Name:                      random.String(128, random.Alphanumeric),
		Description:               random.String(128, random.Alphanumeric),
		KeepaliveTimeoutInSeconds: 15,
		RetryStrategy: &moab.RetryStrategy{
			RetryIntervalsInSeconds: []int64{1, 2, 3, 4, 5},
		},
		DequeuingSettings: &moab.DequeuingSettings{
			MaxInProgressTasks: 100,
			RateLimiting: &moab.TokenBucketRateLimiting{
				MaxTokens:    1000,
				Interval:     5,
				IntervalUnit: moab.IntervalUnit_INTERVAL_UNIT_SECONDS,
			},
			DequeuingPaused: false,
		},
		DeadLetterQueueConfig: &moab.DeadLetterQueueConfig{
			Enable:                   true,
			MaxSize:                  1000000,
			RetentionPeriodInSeconds: 86400 * 14,
		},
		ExpiresInSeconds: 86400,
	}

	// Sign the request
	signature, err := authn.SignAlfa(timestamp, privatePem, request)
	require.NoError(err)

	// Check the signature within 5 minutes time window (timestamp = now)
	err = authn.VerifyAlfaSignature(signature, timestamp, now, publicPem, request)
	require.NoError(err)

	// Check the signature outside 5 minutes time window (timestamp = now + 6 minutes)
	err = authn.VerifyAlfaSignature(signature, timestamp, now.Add(time.Minute*6), publicPem, request)
	require.Error(err)
}

// TestAlfaConsistent tests that Alfa signing mechanism produces the same signature for a given timestamp, request,
// and secret and does not change over time (degradation test). Also, all implementations on different languages should
// produce the same signature given the same input.
func TestAlfaConsistent(t *testing.T) {
	require := require.New(t)

	// Given timestamp
	timestamp := int64(1733240571)
	now := time.Unix(1733240571, 0)

	// Given Alfa key pair
	privatePem := "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIN33cCNGxsuxwMaJ2jWvWcgxBSVr8HV7WUUSKGc71/BtoAoGCCqGSM49\nAwEHoUQDQgAE0m8+ZVijytLp01dsupG7QF8ZpjX5UmP20wj/sluPdoHW3BgiiyCn\n/pMwYptUs0yJUtUZ/0wzEyp8PgAWWhxglw==\n-----END EC PRIVATE KEY-----"
	publicPem := "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE0m8+ZVijytLp01dsupG7QF8ZpjX5\nUmP20wj/sluPdoHW3BgiiyCn/pMwYptUs0yJUtUZ/0wzEyp8PgAWWhxglw==\n-----END PUBLIC KEY-----"

	// Build a fixed request with non-trivial values and nested objects
	request := &moab.CreateQueueRequest{
		Name:                      "7sgaXAYiF7CL1I7PwsaUDjcLpKwRX5oounbJZjI2xcEJDlio6GIPDmn0zPagExSr2Ct0u0eRxSncw4Cf4JmsQbkfFVs722DqJNGG9fGUEsA2JXBi5mBbvih39p8fNhZc",
		Description:               "WSO4ldIQObQKWZQ7fLZuzEX9S9NwXxjIDNaPzw3ktwE6QEASZ3Gw0HZFpgi2Ji9nvqryGbtNH2qWSsN0rrSl27xlF9HF5qnq8rokuZ8fYu2zOiCTx9U2eO3h7WQyqASQ",
		KeepaliveTimeoutInSeconds: 15,
		RetryStrategy: &moab.RetryStrategy{
			RetryIntervalsInSeconds: []int64{1, 2, 3, 4, 5},
		},
		DequeuingSettings: &moab.DequeuingSettings{
			MaxInProgressTasks: 100,
			RateLimiting: &moab.TokenBucketRateLimiting{
				MaxTokens:    1000,
				Interval:     5,
				IntervalUnit: moab.IntervalUnit_INTERVAL_UNIT_SECONDS,
			},
			DequeuingPaused: false,
		},
		DeadLetterQueueConfig: &moab.DeadLetterQueueConfig{
			Enable:                   true,
			MaxSize:                  1000000,
			RetentionPeriodInSeconds: 86400 * 14,
		},
		ExpiresInSeconds: 86400,
	}

	// Sign the request
	signature, err := authn.SignAlfa(timestamp, privatePem, request)
	require.NoError(err)
	// Unfortunately we cannot require that actual signature is equal to some expected value since ECDSA signatures are
	// randomized. Here we only check that it was successfully signed.

	// Signatures from Go SDK
	//signature = "MEQCIHN0cbFwVqwE1Ds++OSIZAjqueLAnENzsNZXqeGTKID0AiBv7Vw2lUyWS9ekCS8s5nItXC35v5O1TFtAaLsvVj+iOw=="

	// Signatures from Ruby SDK
	signature = "MEYCIQCFkp55WmT7lgm+s/mDCxhymP/cGhYEehnpXkFxDawoTgIhAIVpuLVzqik3TkrTqH6aFuMWEUbVZqb1ZIT0tqi3+w5P"
	//signature = "MEYCIQDCBaiHbs46AA5rIkigWChs7RkhrIGr3MbvO10EGipWXQIhAMMAcgpcW65ISe9u3hdBXzSS8J1zhqeGRMzUkpcV6Ila"
	//signature = "MEYCIQDg3GwKSTXTOO23Ywcuz2O8UApVKhmQlc4ASpeyrZPM7QIhAJI2mz90OBtBeKGzg7ZD/Ho9ALOABIPFm78cSLTKH/Y3"

	// Check that a given signature can be successfully verified
	err = authn.VerifyAlfaSignature(signature, timestamp, now, publicPem, request)
	require.NoError(err)
}
