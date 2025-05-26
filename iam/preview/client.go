package iam

import (
	"context"
	"log"

	evrblk "github.com/evrblk/evrblk-go"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type IAMPreviewApi interface {
	CreateRole(ctx context.Context, request *CreateRoleRequest) (*CreateRoleResponse, error)
	GetRole(ctx context.Context, request *GetRoleRequest) (*GetRoleResponse, error)
	UpdateRole(ctx context.Context, request *UpdateRoleRequest) (*UpdateRoleResponse, error)
	ListRoles(ctx context.Context, request *ListRolesRequest) (*ListRolesResponse, error)
	DeleteRole(ctx context.Context, request *DeleteRoleRequest) (*DeleteRoleResponse, error)

	CreateUser(ctx context.Context, request *CreateUserRequest) (*CreateUserResponse, error)
	GetUser(ctx context.Context, request *GetUserRequest) (*GetUserResponse, error)
	UpdateUser(ctx context.Context, request *UpdateUserRequest) (*UpdateUserResponse, error)
	ListUsers(ctx context.Context, request *ListUsersRequest) (*ListUsersResponse, error)
	DeleteUser(ctx context.Context, request *DeleteUserRequest) (*DeleteUserResponse, error)

	CreateApiKey(ctx context.Context, request *CreateApiKeyRequest) (*CreateApiKeyResponse, error)
	GetApiKey(ctx context.Context, request *GetApiKeyRequest) (*GetApiKeyResponse, error)
	ListApiKeys(ctx context.Context, request *ListApiKeysRequest) (*ListApiKeysResponse, error)
	DeleteApiKey(ctx context.Context, request *DeleteApiKeyRequest) (*DeleteApiKeyResponse, error)
}

type IAMPreviewGrpcClient struct {
	grpc IamPreviewApiClient
	conn *grpc.ClientConn

	signer evrblk.RequestSigner
}

var _ IAMPreviewApi = &IAMPreviewGrpcClient{}

func (c *IAMPreviewGrpcClient) WithSigner(signer evrblk.RequestSigner) *IAMPreviewGrpcClient {
	return &IAMPreviewGrpcClient{
		grpc:   c.grpc,
		conn:   c.conn,
		signer: signer,
	}
}

func (c *IAMPreviewGrpcClient) CreateRole(ctx context.Context, request *CreateRoleRequest) (*CreateRoleResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.CreateRole(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) GetRole(ctx context.Context, request *GetRoleRequest) (*GetRoleResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.GetRole(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) UpdateRole(ctx context.Context, request *UpdateRoleRequest) (*UpdateRoleResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.UpdateRole(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) ListRoles(ctx context.Context, request *ListRolesRequest) (*ListRolesResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.ListRoles(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) DeleteRole(ctx context.Context, request *DeleteRoleRequest) (*DeleteRoleResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.DeleteRole(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) CreateUser(ctx context.Context, request *CreateUserRequest) (*CreateUserResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.CreateUser(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) GetUser(ctx context.Context, request *GetUserRequest) (*GetUserResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.GetUser(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) UpdateUser(ctx context.Context, request *UpdateUserRequest) (*UpdateUserResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.UpdateUser(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) ListUsers(ctx context.Context, request *ListUsersRequest) (*ListUsersResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.ListUsers(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) DeleteUser(ctx context.Context, request *DeleteUserRequest) (*DeleteUserResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.DeleteUser(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) CreateApiKey(ctx context.Context, request *CreateApiKeyRequest) (*CreateApiKeyResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.CreateApiKey(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) GetApiKey(ctx context.Context, request *GetApiKeyRequest) (*GetApiKeyResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.GetApiKey(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) ListApiKeys(ctx context.Context, request *ListApiKeysRequest) (*ListApiKeysResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.ListApiKeys(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) DeleteApiKey(ctx context.Context, request *DeleteApiKeyRequest) (*DeleteApiKeyResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.DeleteApiKey(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *IAMPreviewGrpcClient) Close() {
	c.conn.Close()
}

func NewIAMPreviewGrpcClient(address string, signer evrblk.RequestSigner) *IAMPreviewGrpcClient {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return &IAMPreviewGrpcClient{
		conn:   conn,
		grpc:   NewIamPreviewApiClient(conn),
		signer: signer,
	}
}
