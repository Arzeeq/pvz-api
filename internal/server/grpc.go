package server

import (
	"context"
	"errors"

	pb "github.com/Arzeeq/pvz-api/internal/grpc"
	"google.golang.org/grpc"
)

type GrpcHandler interface {
	GetPVZList(ctx context.Context, req *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error)
}

type GRPCServer struct {
	pb.UnimplementedPVZServiceServer
	handler GrpcHandler
}

func (s *GRPCServer) GetPVZList(ctx context.Context, req *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error) {
	return s.handler.GetPVZList(ctx, req)
}

func NewGRPC(handler GrpcHandler) (*grpc.Server, error) {
	if handler == nil {
		return nil, errors.New("nil values in constructor")
	}

	s := grpc.NewServer()
	pb.RegisterPVZServiceServer(s, &GRPCServer{handler: handler})
	return s, nil
}
