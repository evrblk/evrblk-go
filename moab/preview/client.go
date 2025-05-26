package moab

import (
	"context"
	"log"

	evrblk "github.com/evrblk/evrblk-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MoabPreviewApi interface {
	CreateQueue(ctx context.Context, request *CreateQueueRequest) (*CreateQueueResponse, error)
	GetQueue(ctx context.Context, request *GetQueueRequest) (*GetQueueResponse, error)
	UpdateQueue(ctx context.Context, request *UpdateQueueRequest) (*UpdateQueueResponse, error)
	DeleteQueue(ctx context.Context, request *DeleteQueueRequest) (*DeleteQueueResponse, error)
	ListQueues(ctx context.Context, request *ListQueuesRequest) (*ListQueuesResponse, error)

	CreateSchedule(ctx context.Context, request *CreateScheduleRequest) (*CreateScheduleResponse, error)
	GetSchedule(ctx context.Context, request *GetScheduleRequest) (*GetScheduleResponse, error)
	UpdateSchedule(ctx context.Context, request *UpdateScheduleRequest) (*UpdateScheduleResponse, error)
	DeleteSchedule(ctx context.Context, request *DeleteScheduleRequest) (*DeleteScheduleResponse, error)

	GetTask(ctx context.Context, request *GetTaskRequest) (*GetTaskResponse, error)
	Enqueue(ctx context.Context, request *EnqueueRequest) (*EnqueueResponse, error)
	Dequeue(ctx context.Context, request *DequeueRequest) (*DequeueResponse, error)
	ReportStatus(ctx context.Context, request *ReportStatusRequest) (*ReportStatusResponse, error)
	DeleteTasks(ctx context.Context, request *DeleteTasksRequest) (*DeleteTasksResponse, error)
	RestartTasks(ctx context.Context, request *RestartTasksRequest) (*RestartTasksResponse, error)
	PurgeQueue(ctx context.Context, request *PurgeQueueRequest) (*PurgeQueueResponse, error)
}

type MoabPreviewGrpcClient struct {
	grpc MoabPreviewApiClient
	conn *grpc.ClientConn

	signer evrblk.RequestSigner
}

func (c *MoabPreviewGrpcClient) WithSigner(signer evrblk.RequestSigner) *MoabPreviewGrpcClient {
	return &MoabPreviewGrpcClient{
		grpc:   c.grpc,
		conn:   c.conn,
		signer: signer,
	}
}

var _ MoabPreviewApi = &MoabPreviewGrpcClient{}

func (c *MoabPreviewGrpcClient) CreateQueue(ctx context.Context, request *CreateQueueRequest) (*CreateQueueResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.CreateQueue(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) GetQueue(ctx context.Context, request *GetQueueRequest) (*GetQueueResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.GetQueue(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) UpdateQueue(ctx context.Context, request *UpdateQueueRequest) (*UpdateQueueResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.UpdateQueue(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) DeleteQueue(ctx context.Context, request *DeleteQueueRequest) (*DeleteQueueResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.DeleteQueue(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) ListQueues(ctx context.Context, request *ListQueuesRequest) (*ListQueuesResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.ListQueues(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) CreateSchedule(ctx context.Context, request *CreateScheduleRequest) (*CreateScheduleResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.CreateSchedule(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) GetSchedule(ctx context.Context, request *GetScheduleRequest) (*GetScheduleResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.GetSchedule(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) UpdateSchedule(ctx context.Context, request *UpdateScheduleRequest) (*UpdateScheduleResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.UpdateSchedule(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) DeleteSchedule(ctx context.Context, request *DeleteScheduleRequest) (*DeleteScheduleResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.DeleteSchedule(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) GetTask(ctx context.Context, request *GetTaskRequest) (*GetTaskResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.GetTask(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) Enqueue(ctx context.Context, request *EnqueueRequest) (*EnqueueResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.Enqueue(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) Dequeue(ctx context.Context, request *DequeueRequest) (*DequeueResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.Dequeue(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) ReportStatus(ctx context.Context, request *ReportStatusRequest) (*ReportStatusResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.ReportStatus(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) DeleteTasks(ctx context.Context, request *DeleteTasksRequest) (*DeleteTasksResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.DeleteTasks(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) RestartTasks(ctx context.Context, request *RestartTasksRequest) (*RestartTasksResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.RestartTasks(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) PurgeQueue(ctx context.Context, request *PurgeQueueRequest) (*PurgeQueueResponse, error) {
	signedCtx, err := c.signer.Sign(ctx, request)
	if err != nil {
		return nil, err
	}

	resp, err := c.grpc.PurgeQueue(signedCtx, request)

	return resp, evrblk.FromRpcError(err)
}

func (c *MoabPreviewGrpcClient) Close() {
	c.conn.Close()
}

func NewMoabPreviewGrpcClient(address string, signer evrblk.RequestSigner) *MoabPreviewGrpcClient {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return &MoabPreviewGrpcClient{
		conn:   conn,
		grpc:   NewMoabPreviewApiClient(conn),
		signer: signer,
	}
}
