package client

import (
	"context"
	"log"
	"time"

	pb "github.com/hokhoa05/core-engine/pb/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SettlementClient struct {
	client pb.SettlementEngineClient
}

func NewSettlementClient(address string) (*SettlementClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &SettlementClient{
		client: pb.NewSettlementEngineClient(conn),
	}, nil
}

func (s *SettlementClient) ReportTrade(makerOrderID, takerOrderID, makerUserID, takerUserID, price, qty uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.TradeReportRequest{
		MakerOrderId: int64(makerOrderID),
		TakerOrderId: int64(takerOrderID),
		MakerUserId:  int64(makerUserID),
		TakerUserId:  int64(takerUserID),
		Price:        int64(price),
		Qty:          int64(qty),
	}

	_, err := s.client.ReportTrade(ctx, req)
	if err != nil {
		log.Printf("Failed when settlement for trade (Maker: %d, Taker: %d): %v", makerOrderID, takerOrderID, err)
		return err
	}

	log.Printf("Successfully sent Settlement to Java -> Trade(Price: %d, Qty: %d)", price, qty)
	return nil
}
