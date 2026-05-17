package providers

import (
	"context"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	graderPB "github.com/CSKU-Lab/main-server/genproto/grader/v1"
	taskPB "github.com/CSKU-Lab/main-server/genproto/task/v1"
	"github.com/CSKU-Lab/main-server/internal/adapters/pubsub"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest"
	"github.com/CSKU-Lab/main-server/internal/adapters/storage/minio"
	"github.com/CSKU-Lab/queue"
	"github.com/google/wire"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitConfigGRPCClient(clientAddr string) (configPB.ConfigServiceClient, func(), error) {
	conn, err := grpc.NewClient(clientAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, nil, err
	}
	client := configPB.NewConfigServiceClient(conn)
	return client, func() { conn.Close() }, nil
}

func InitGraderGRPCClient(clientAddr string) (graderPB.GraderServiceClient, func(), error) {
	conn, err := grpc.NewClient(clientAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, nil, err
	}
	client := graderPB.NewGraderServiceClient(conn)
	return client, func() { conn.Close() }, nil
}

func InitTaskGRPCClient(clientAddr string) (taskPB.TaskServiceClient, func(), error) {
	conn, err := grpc.NewClient(clientAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, nil, err
	}
	client := taskPB.NewTaskServiceClient(conn)
	return client, func() { conn.Close() }, nil
}

func ProvideGraderClient(cfg *configs.Config) (graderPB.GraderServiceClient, func(), error) {
	return InitGraderGRPCClient(cfg.GRADER_SERVER_URL)
}

func ProvideConfigClient(cfg *configs.Config) (configPB.ConfigServiceClient, func(), error) {
	return InitConfigGRPCClient(cfg.CONFIG_SERVER_URL)
}

func ProvideTaskClient(cfg *configs.Config) (taskPB.TaskServiceClient, func(), error) {
	return InitTaskGRPCClient(cfg.TASK_SERVER_URL)
}

func ProvideRedis(cfg *configs.Config) (pubsub.PubSub, error) {
	return pubsub.NewRedis(cfg.REDIS_ADDR, cfg.REDIS_PASSWORD)
}

func ProvideMinio(ctx context.Context, cfg *configs.Config) repositories.FileRepository {
	return minio.New(ctx, cfg)
}

func ProvideRabbitMQ(cfg *configs.Config) (queue.Queue, error) {
	return queue.NewRabbitMQ(cfg.RBMQ_SERVER_URL)
}

func NewPlaygroundHandler(graderClient graderPB.GraderServiceClient) *rest.PlaygroundHandler {
	return rest.NewPlaygroundHandler(graderClient)
}

var ExternalSet = wire.NewSet(
	ProvideGraderClient,
	ProvideConfigClient,
	ProvideTaskClient,
	ProvideRedis,
	ProvideMinio,
	ProvideRabbitMQ,
	NewPlaygroundHandler,
)
