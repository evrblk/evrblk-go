package grackle

import (
	"context"
	"log"

	evrblk "github.com/evrblk/evrblk-go"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GracklePreviewApi interface {
	CreateNamespace(ctx context.Context, request *CreateNamespaceRequest) (*CreateNamespaceResponse, error)
	GetNamespace(ctx context.Context, request *GetNamespaceRequest) (*GetNamespaceResponse, error)
	UpdateNamespace(ctx context.Context, request *UpdateNamespaceRequest) (*UpdateNamespaceResponse, error)
	DeleteNamespace(ctx context.Context, request *DeleteNamespaceRequest) (*DeleteNamespaceResponse, error)
	ListNamespaces(ctx context.Context, request *ListNamespacesRequest) (*ListNamespacesResponse, error)

	CreateSemaphore(ctx context.Context, request *CreateSemaphoreRequest) (*CreateSemaphoreResponse, error)
	GetSemaphore(ctx context.Context, request *GetSemaphoreRequest) (*GetSemaphoreResponse, error)
	UpdateSemaphore(ctx context.Context, request *UpdateSemaphoreRequest) (*UpdateSemaphoreResponse, error)
	DeleteSemaphore(ctx context.Context, request *DeleteSemaphoreRequest) (*DeleteSemaphoreResponse, error)

	CreateWaitGroup(ctx context.Context, request *CreateWaitGroupRequest) (*CreateWaitGroupResponse, error)
	GetWaitGroup(ctx context.Context, request *GetWaitGroupRequest) (*GetWaitGroupResponse, error)
	DeleteWaitGroup(ctx context.Context, request *DeleteWaitGroupRequest) (*DeleteWaitGroupResponse, error)
	AddJobToWaitGroup(ctx context.Context, request *AddJobToWaitGroupRequest) (*AddJobToWaitGroupResponse, error)
	CompleteJobFromWaitGroup(ctx context.Context, request *CompleteJobFromWaitGroupRequest) (*CompleteJobFromWaitGroupResponse, error)

	AcquireLock(ctx context.Context, request *AcquireLockRequest) (*AcquireLockResponse, error)
	ReleaseLock(ctx context.Context, request *ReleaseLockRequest) (*ReleaseLockResponse, error)
	GetLock(ctx context.Context, request *GetLockRequest) (*GetLockResponse, error)
	DeleteLock(ctx context.Context, request *DeleteLockRequest) (*DeleteLockResponse, error)
}

type GracklePreviewGrpcClient struct {
	grpc GracklePreviewApiClient
	conn *grpc.ClientConn

	signer evrblk.RequestSigner
}

func (c *GracklePreviewGrpcClient) WithSigner(signer evrblk.RequestSigner) *GracklePreviewGrpcClient {
	return &GracklePreviewGrpcClient{
		grpc:   c.grpc,
		conn:   c.conn,
		signer: signer,
	}
}

var _ GracklePreviewApi = &GracklePreviewGrpcClient{}

func (c *GracklePreviewGrpcClient) CreateNamespace(ctx context.Context, request *CreateNamespaceRequest) (*CreateNamespaceResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.CreateNamespace(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) GetNamespace(ctx context.Context, request *GetNamespaceRequest) (*GetNamespaceResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.GetNamespace(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) UpdateNamespace(ctx context.Context, request *UpdateNamespaceRequest) (*UpdateNamespaceResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.UpdateNamespace(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) DeleteNamespace(ctx context.Context, request *DeleteNamespaceRequest) (*DeleteNamespaceResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.DeleteNamespace(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) ListNamespaces(ctx context.Context, request *ListNamespacesRequest) (*ListNamespacesResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.ListNamespaces(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) CreateSemaphore(ctx context.Context, request *CreateSemaphoreRequest) (*CreateSemaphoreResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.CreateSemaphore(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) GetSemaphore(ctx context.Context, request *GetSemaphoreRequest) (*GetSemaphoreResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.GetSemaphore(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) UpdateSemaphore(ctx context.Context, request *UpdateSemaphoreRequest) (*UpdateSemaphoreResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.UpdateSemaphore(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) DeleteSemaphore(ctx context.Context, request *DeleteSemaphoreRequest) (*DeleteSemaphoreResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.DeleteSemaphore(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) CreateWaitGroup(ctx context.Context, request *CreateWaitGroupRequest) (*CreateWaitGroupResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.CreateWaitGroup(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) GetWaitGroup(ctx context.Context, request *GetWaitGroupRequest) (*GetWaitGroupResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.GetWaitGroup(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) DeleteWaitGroup(ctx context.Context, request *DeleteWaitGroupRequest) (*DeleteWaitGroupResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.DeleteWaitGroup(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) AddJobToWaitGroup(ctx context.Context, request *AddJobToWaitGroupRequest) (*AddJobToWaitGroupResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.AddJobToWaitGroup(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) CompleteJobFromWaitGroup(ctx context.Context, request *CompleteJobFromWaitGroupRequest) (*CompleteJobFromWaitGroupResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.CompleteJobFromWaitGroup(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) AcquireLock(ctx context.Context, request *AcquireLockRequest) (*AcquireLockResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.AcquireLock(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) ReleaseLock(ctx context.Context, request *ReleaseLockRequest) (*ReleaseLockResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.ReleaseLock(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) GetLock(ctx context.Context, request *GetLockRequest) (*GetLockResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.GetLock(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) DeleteLock(ctx context.Context, request *DeleteLockRequest) (*DeleteLockResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.DeleteLock(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *GracklePreviewGrpcClient) Close() {
	c.conn.Close()
}

func NewGracklePreviewGrpcClient(address string, signer evrblk.RequestSigner) *GracklePreviewGrpcClient {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return &GracklePreviewGrpcClient{
		conn:   conn,
		grpc:   NewGracklePreviewApiClient(conn),
		signer: signer,
	}
}
