package mcar

import (
	"context"

	"github.com/cloud-barista/cb-ladybug/src/grpc-api/logger"
	pb "github.com/cloud-barista/cb-ladybug/src/grpc-api/protobuf/cbladybug"
)

// ===== [ Constants and Variables ] =====

// ===== [ Types ] =====

// ===== [ Implementations ] =====

// Healthy - 상태확인
func (s *MCARService) Healthy(ctx context.Context, req *pb.Empty) (*pb.MessageResponse, error) {
	logger := logger.NewLogger()

	logger.Debug("calling MCARService.Healthy()")

	resp := &pb.MessageResponse{Message: "cb-barista cb-ladybug"}
	return resp, nil
}

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====
