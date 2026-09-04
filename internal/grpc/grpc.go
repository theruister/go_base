package grpc

import (
	"context"
	"sync"
	"time"

	basehdlr "bitbucket.org/theruister/goBase/internal/basehdlr"
	grpc "bitbucket.org/theruister/goBase/pkg/api/v1/gen/go"
)

type Server struct {
	mu    *sync.RWMutex
	users map[string]*grpc.AddUserRequest
	grpc.UnimplementedGoBaseServiceServer
}

type ServiceStatus struct {
	StartTime        time.Time
	TransactionCount int
	GrpcError        bool
	GrpcMessage      string
	DbError          bool
	DbMessage        string
	MqError          bool
	MqMessage        string
}

var Status = newStatus()

func newStatus() ServiceStatus {
	return ServiceStatus{StartTime: time.Now(), TransactionCount: 0, GrpcError: false, GrpcMessage: "", DbError: false, DbMessage: "", MqError: false, MqMessage: ""}
}

func GetServiceStatus() *ServiceStatus {
	return &Status
}

func New() *Server {
	return &Server{
		mu:    &sync.RWMutex{},
		users: make(map[string]*grpc.AddUserRequest),
	}
}

func (s *Server) AddUser(ctx context.Context, user *grpc.AddUserRequest) (*grpc.AddUserResponse, error) {
	return basehdlr.AddUser(ctx, user)
}

func (s *Server) GetUsers(ctx context.Context, user *grpc.AddUserRequest) (*grpc.AddUserResponse, error) {
	return basehdlr.GetUsers(ctx, user)
}
