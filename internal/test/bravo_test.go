package test

import (
	"testing"
	"time"

	"github.com/evrblk/evrblk-go/authn"
	moab "github.com/evrblk/evrblk-go/moab/preview"
	"github.com/labstack/gommon/random"
	"github.com/stretchr/testify/require"
)

// TestBravoSignAndVerify tests Bravo signing mechanism on a random timestamp, request and secret. A signature should be
// valid within 5 minutes time window.
func TestBravoSignAndVerify(t *testing.T) {
	// Get current time
	now := time.Now()
	timestamp := now.Unix()

	// Generate a new Bravo secret string
	secret := authn.GenerateBravoSecret()

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
	signature, err := authn.SignBravo(timestamp, secret, request, "Moab", "CreateQueueRequest")
	require.NoError(t, err)

	// Get hashed secret
	date := authn.GetDateOfTimestamp(timestamp)
	hashedSecret, err := authn.HashBravoSecretWithDate(secret, date)
	require.NoError(t, err)

	// Check the signature within 5 minutes time window (timestamp = now)
	err = authn.VerifyBravoSignature(signature, timestamp, now, hashedSecret, request, "Moab", "CreateQueueRequest")
	require.NoError(t, err)

	// Check the signature outside 5 minutes time window (timestamp = now + 6 minutes)
	err = authn.VerifyBravoSignature(signature, timestamp, now.Add(time.Minute*6), hashedSecret, request, "Moab", "CreateQueueRequest")
	require.Error(t, err)

	// Check the signature for different service name
	err = authn.VerifyBravoSignature(signature, timestamp, now.Add(time.Minute*6), hashedSecret, request, "Jakal", "CreateQueueRequest")
	require.Error(t, err)

	// Check the signature for different method name
	err = authn.VerifyBravoSignature(signature, timestamp, now.Add(time.Minute*6), hashedSecret, request, "Moab", "DeleteQueueRequest")
	require.Error(t, err)
}

// TestBravoConsistent tests that Bravo signing mechanism produces the same signature for a given timestamp, request,
// and secret and does not change over time (degradation test). Also, all implementations on different languages should
// produce the same signature given the same input.
func TestBravoConsistent(t *testing.T) {
	// Given timestamp
	timestamp := int64(1733240571)
	now := time.Unix(1733240571, 0)

	// Given secret
	secret := "yuOf1mvP9el6iuTz+UqhDdAlHD5o1hKMY7LF2mB5pdhSGftrlZZ7zU1RURBIqSpKIWRVaB7u/L2RmLcbIcOxaxKgHYCcGcnAc6BHCoeWFCahiOKO3zpvA5ebsEoMlFVaPN+JtDTwijJuyIzG8Uus6w2pK7aleEUqvUKzdibx19u6TKFmrZ4GqpzkrLt7HSEAIn8rZaIHxJ8HKQYtbwBRknEMs55n9oUGf3hYTB8TYbydbIByV35HB+uTyxbmNibUC+4khKQPRtNmHR5/fKMUdk7tZvGkmmAktNCfBOz9mrmpJao75bY9KMwTq5k4y8I0ZbtY4RvvT2RXIMYA/RInDlk647B3/XqsR9uGtEqoMrTu3ZaEWUpMShSoSoWqLBxb3KIzeIEwnmy1lA9KMQ2HaLVwN9vrndN+x76vyNXBarLfakGP9pDGrYBxWxMW7pUptLQ81iJUkhFt/fv3JelPPxeqo96zmU8cPcfdXIwtqnLzd3o/P3DSbv9f+ubIc5Wt0DDFIWLg2lgpUmpF22MWZ9YZR/D/2cidOHTue+XG+N3OPqEhI6JNKJjWQ46pPwm/uLVIhIw58PhHnC5yLs58xpRp817WJ28RS58vCLNZ0ddG9RrZkHWv8MEd97DAWZHp9+1kGlerSZytiohGPPJWngS5t6Z+JySbNFvXuXpmrs4="

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
	signature, err := authn.SignBravo(timestamp, secret, request, "Moab", "CreateQueueRequest")
	require.NoError(t, err)
	// The same request, same timestamp, and same secret should always produce the same signature
	require.Equal(t, "4b84547a1dd982dec9ffcea97940b92b48b00f3bf1be87336189cbcbc7f1f50c", signature)

	// Get hashed secret
	date := authn.GetDateOfTimestamp(timestamp)
	hashedSecret, err := authn.HashBravoSecretWithDate(secret, date)
	require.NoError(t, err)

	// Check that this signature can be successfully verified
	err = authn.VerifyBravoSignature(signature, timestamp, now, hashedSecret, request, "Moab", "CreateQueueRequest")
	require.NoError(t, err)

	// Check that this signature can not be verified for another service
	err = authn.VerifyBravoSignature(signature, timestamp, now, hashedSecret, request, "Jakal", "CreateQueueRequest")
	require.Error(t, err)
}
