package evrblk

import (
	"context"
	"fmt"
	"time"

	"github.com/evrblk/evrblk-go/authn"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

const (
	signatureKey = "evrblk-signature"
	apiKeyKey    = "evrblk-api-key-id"
	timestampKey = "evrblk-timestamp"
)

type RequestSigner interface {
	Sign(ctx context.Context, request proto.Message) (context.Context, error)
}

type alfaRequestSigner struct {
	privatePem string
	apiKeyId   string
}

var _ RequestSigner = &alfaRequestSigner{}

func (s *alfaRequestSigner) Sign(ctx context.Context, request proto.Message) (context.Context, error) {
	// Current time in Unix seconds
	now := time.Now().Unix()

	signature, err := authn.SignAlfa(now, s.privatePem, request)
	if err != nil {
		return nil, err
	}

	// Add headers into request context
	ctx = metadata.AppendToOutgoingContext(ctx,
		apiKeyKey, s.apiKeyId, // API key ID
		timestampKey, fmt.Sprintf("%d", now), // Current timestamp
		signatureKey, signature) // Signature

	return ctx, nil
}

func NewAlfaRequestSigner(apiKeyId string, privatePem string) (RequestSigner, error) {
	// TODO add validations: key is alfa, pem is valid
	return &alfaRequestSigner{
		privatePem: privatePem,
		apiKeyId:   apiKeyId,
	}, nil
}

type bravoRequestSigner struct {
	secret   string
	apiKeyId string
}

var _ RequestSigner = &bravoRequestSigner{}

func (s *bravoRequestSigner) Sign(ctx context.Context, request proto.Message) (context.Context, error) {
	// Current time in Unix seconds
	now := time.Now().Unix()

	signature, err := authn.SignBravo(now, s.secret, request)
	if err != nil {
		return nil, err
	}

	// Add headers into request context
	ctx = metadata.AppendToOutgoingContext(ctx,
		apiKeyKey, s.apiKeyId, // API key ID
		timestampKey, fmt.Sprintf("%d", now), // Current timestamp
		signatureKey, signature) // Signature

	return ctx, nil
}

func NewBravoRequestSigner(apiKeyId string, apiSecretKey string) (RequestSigner, error) {
	// TODO add validations: key is bravo, secret is valid
	return &bravoRequestSigner{
		secret:   apiSecretKey,
		apiKeyId: apiKeyId,
	}, nil
}
