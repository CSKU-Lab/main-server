//go:build e2e
// +build e2e

package e2e

import (
	"context"

	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	taskPB "github.com/CSKU-Lab/main-server/genproto/task/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TaskServiceStub is a stub implementation of taskPB.TaskServiceClient for testing
type TaskServiceStub struct{}

func NewTaskServiceStub() taskPB.TaskServiceClient {
	return &TaskServiceStub{}
}

func (s *TaskServiceStub) GetTasks(ctx context.Context, in *taskPB.GetTasksRequest, opts ...grpc.CallOption) (*taskPB.GetTasksResponse, error) {
	return &taskPB.GetTasksResponse{
		Tasks: []*taskPB.TaskResponse{},
	}, nil
}

func (s *TaskServiceStub) GetTask(ctx context.Context, in *taskPB.GetTaskRequest, opts ...grpc.CallOption) (*taskPB.TaskResponse, error) {
	compareScriptId := "test-compare-id"
	return &taskPB.TaskResponse{
		Id:              in.Id,
		TestCaseGroups:  []*taskPB.TestCaseGroup{},
		AllowedRunners:  []*taskPB.AllowedRunner{},
		CompareScriptId: &compareScriptId,
	}, nil
}

func (s *TaskServiceStub) CreateTask(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*taskPB.CreateTaskResponse, error) {
	return &taskPB.CreateTaskResponse{
		Id: "test-task-id-" + randomString(8),
	}, nil
}

func (s *TaskServiceStub) UpdateTask(ctx context.Context, in *taskPB.UpdateTaskRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *TaskServiceStub) DeleteTask(ctx context.Context, in *taskPB.DeleteTaskRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *TaskServiceStub) RemoveRunnerOnCascade(ctx context.Context, in *taskPB.RemoveRunnerOnCascadeRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *TaskServiceStub) RemoveCompareScriptOnCascade(ctx context.Context, in *taskPB.RemoveCompareScriptOnCascadeRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// ConfigServiceStub is a stub implementation of configPB.ConfigServiceClient for testing
type ConfigServiceStub struct{}

func NewConfigServiceStub() configPB.ConfigServiceClient {
	return &ConfigServiceStub{}
}

func (s *ConfigServiceStub) CreateRunner(ctx context.Context, in *configPB.CreateRunnerRequest, opts ...grpc.CallOption) (*configPB.CreateRunnerResponse, error) {
	return &configPB.CreateRunnerResponse{
		Id: "test-runner-id-" + randomString(8),
	}, nil
}

func (s *ConfigServiceStub) GetRunnersPagination(ctx context.Context, in *configPB.GetRunnersPaginationRequest, opts ...grpc.CallOption) (*configPB.GetRunnersPaginationResponse, error) {
	return &configPB.GetRunnersPaginationResponse{
		Runners: []*configPB.RunnerPaginationData{},
	}, nil
}

func (s *ConfigServiceStub) GetRunner(ctx context.Context, in *configPB.GetRunnerRequest, opts ...grpc.CallOption) (*configPB.RunnerResponse, error) {
	return &configPB.RunnerResponse{
		Id:   in.Id,
		Name: "Test Runner",
	}, nil
}

func (s *ConfigServiceStub) UpdateRunner(ctx context.Context, in *configPB.UpdateRunnerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *ConfigServiceStub) DeleteRunner(ctx context.Context, in *configPB.DeleteRunnerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *ConfigServiceStub) GetAllRunners(ctx context.Context, in *configPB.GetAllRunnersRequest, opts ...grpc.CallOption) (*configPB.GetAllRunnersResponse, error) {
	return &configPB.GetAllRunnersResponse{
		Runners: []*configPB.RunnerResponse{},
	}, nil
}

func (s *ConfigServiceStub) CreateCompare(ctx context.Context, in *configPB.CreateCompareRequest, opts ...grpc.CallOption) (*configPB.CreateCompareResponse, error) {
	return &configPB.CreateCompareResponse{
		Id: "test-compare-id-" + randomString(8),
	}, nil
}

func (s *ConfigServiceStub) GetComparesPagination(ctx context.Context, in *configPB.GetComparesPaginationRequest, opts ...grpc.CallOption) (*configPB.GetComparesPaginationResponse, error) {
	return &configPB.GetComparesPaginationResponse{
		Compares: []*configPB.CompareResponse{},
	}, nil
}

func (s *ConfigServiceStub) GetCompare(ctx context.Context, in *configPB.GetCompareRequest, opts ...grpc.CallOption) (*configPB.CompareResponse, error) {
	return &configPB.CompareResponse{
		Id:   in.Id,
		Name: "Test Compare Script",
	}, nil
}

func (s *ConfigServiceStub) GetAllCompares(ctx context.Context, in *configPB.GetAllComparesRequest, opts ...grpc.CallOption) (*configPB.GetAllComparesResponse, error) {
	return &configPB.GetAllComparesResponse{
		Compares: []*configPB.CompareResponse{},
	}, nil
}

func (s *ConfigServiceStub) UpdateCompare(ctx context.Context, in *configPB.UpdateCompareRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *ConfigServiceStub) DeleteCompare(ctx context.Context, in *configPB.DeleteCompareRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// randomString generates a random string of given length (helper function)
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
