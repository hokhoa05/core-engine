package main

import (
	"fmt"
	"log"

	"github.com/hokhoa05/core-engine/internal/client"
	"github.com/hokhoa05/core-engine/internal/matching"
	"github.com/hokhoa05/core-engine/internal/server"
	"github.com/hokhoa05/core-engine/internal/worker"
)

func main() {
	fmt.Println("Core Engine: Distributed Order Matching System is initializing...")
	settlementClient, err := client.NewSettlementClient("localhost:9090")
	if err != nil {
		log.Fatalf("Cannot connect to Java Settlement Server: %v", err)
	}
	runner := matching.NewEngineRunner(10000, 10000)
	go runner.Start()

	settlementWorker := worker.NewSettlementWorker(settlementClient, runner.TradeChannel())
	go settlementWorker.Start()

	server.StartGRPCServer(":50051", runner)
}
