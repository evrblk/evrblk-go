package evrblk

import (
	"context"
	"fmt"
	"strings"
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
	Sign(ctx context.Context, request proto.Message, service string, method string) (context.Context, error)
}

type alfaRequestSigner struct {
	privatePem string
	apiKeyId   string
}

var _ RequestSigner = &alfaRequestSigner{}

func (s *alfaRequestSigner) Sign(ctx context.Context, request proto.Message, service string, method string) (context.Context, error) {
	// Current time in Unix seconds
	now := time.Now().Unix()

	signature, err := authn.SignAlfa(now, s.privatePem, request, service, method)
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

// NewAlfaRequestSigner creates a new request signer for Alfa API keys.
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

func (s *bravoRequestSigner) Sign(ctx context.Context, request proto.Message, service string, method string) (context.Context, error) {
	// Current time in Unix seconds
	now := time.Now().Unix()

	signature, err := authn.SignBravo(now, s.secret, request, service, method)
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

// NewBravoRequestSigner creates a new request signer for Bravo API keys.
func NewBravoRequestSigner(apiKeyId string, apiSecretKey string) (RequestSigner, error) {
	// TODO add validations: key is bravo, secret is valid
	return &bravoRequestSigner{
		secret:   apiSecretKey,
		apiKeyId: apiKeyId,
	}, nil
}

// NewRequestSigner creates a new request signer for Alfa or Bravo API keys based on provided API key ID.
func NewRequestSigner(apiKeyId string, apiSecretKey string) (RequestSigner, error) {
	if strings.HasPrefix(apiKeyId, "key_alfa_") {
		return NewAlfaRequestSigner(apiKeyId, apiSecretKey)
	} else {
		return NewBravoRequestSigner(apiKeyId, apiSecretKey)
	}
}

type noOpSigner struct {
}

var _ RequestSigner = &noOpSigner{}

func (s *noOpSigner) Sign(ctx context.Context, request proto.Message, service string, method string) (context.Context, error) {
	return ctx, nil
}

// NewNoOpSigner creates a new NoOp request signer
func NewNoOpSigner() RequestSigner {
	return &noOpSigner{}
}
