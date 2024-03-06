package grpclib

import (
	"context"
	"github.com/lquyet/distributed-lock-demo/server/pb"
)

// HealthServer ...
type HealthServer struct {
	pb.UnimplementedHealthCheckServiceServer
}

// NewHealthServer ...
func NewHealthServer() *HealthServer {
	return &HealthServer{}
}

// Liveness ...
func (s *HealthServer) Liveness(_ context.Context, _ *pb.LivenessRequest) (*pb.LivenessResponse, error) {
	return &pb.LivenessResponse{}, nil
}

// Readiness ...
func (s *HealthServer) Readiness(_ context.Context, _ *pb.ReadinessRequest) (*pb.ReadinessResponse, error) {
	return &pb.ReadinessResponse{}, nil
}
