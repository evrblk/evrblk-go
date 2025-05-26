package myaccount

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	evrblk "github.com/evrblk/evrblk-go"
)

type MyAccountPreviewApi interface {
	GetAccount(ctx context.Context, request *GetAccountRequest) (*GetAccountResponse, error)
}

type MyAccountPreviewGrpcClient struct {
	grpc MyAccountPreviewApiClient
	conn *grpc.ClientConn

	signer evrblk.RequestSigner
}

var _ MyAccountPreviewApi = &MyAccountPreviewGrpcClient{}

func (c *MyAccountPreviewGrpcClient) WithSigner(signer evrblk.RequestSigner) *MyAccountPreviewGrpcClient {
	return &MyAccountPreviewGrpcClient{
		grpc:   c.grpc,
		conn:   c.conn,
		signer: signer,
	}
}

func (c *MyAccountPreviewGrpcClient) GetAccount(ctx context.Context, request *GetAccountRequest) (*GetAccountResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.GetAccount(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MyAccountPreviewGrpcClient) Close() {
	c.conn.Close()
}

func NewMyAccountPreviewGrpcClient(address string, signer evrblk.RequestSigner) *MyAccountPreviewGrpcClient {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return &MyAccountPreviewGrpcClient{
		conn:   conn,
		grpc:   NewMyAccountPreviewApiClient(conn),
		signer: signer,
	}
}
