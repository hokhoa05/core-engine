package server

import (
	"context"
	"log"
	"net"

	"github.com/hokhoa05/core-engine/internal/matching"
	"github.com/hokhoa05/core-engine/internal/models"
	pb "github.com/hokhoa05/core-engine/pb/proto"
	"google.golang.org/grpc"
)

type grpcServer struct {
	pb.UnimplementedMatchingEngineServer
	runner *matching.EngineRunner
}

func NewGrpcServer(runner *matching.EngineRunner) *grpcServer {
	return &grpcServer{
		runner: runner,
	}
}

func (s *grpcServer) PushOrder(ctx context.Context, req *pb.OrderRequest) (*pb.OrderResponse, error) {
	log.Printf("[gRPC] Received Order: %d from User: %d", req.OrderId, req.UserId)

	side := models.Buy
	if req.Side == "SELL" {
		side = models.Sell
	}

	cmd := matching.Command{
		Type: matching.CmdPlaceOrder,
		Order: models.Order{
			ID:     uint64(req.OrderId),
			UserID: uint64(req.UserId),
			Side:   side,
			Price:  uint64(req.Price),
			Qty:    uint64(req.Qty),
		},
	}

	s.runner.PushCommand(cmd)

	return &pb.OrderResponse{
		Success: true,
		Message: "Order accepted by Core Engine",
	}, nil
}

func StartGRPCServer(port string, runner *matching.EngineRunner) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Cannot open port: %s: %v", port, err)
	}
	s := grpc.NewServer()
	pb.RegisterMatchingEngineServer(s, NewGrpcServer(runner))

	log.Printf("gRPC Server is listening to port %s...", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC Server down: %v", err)
	}
}
