package main

import (
	"fmt"

	"github.com/hokhoa05/core-engine/internal/matching"
	"github.com/hokhoa05/core-engine/internal/server"
)

func main() {
	fmt.Println("Core Engine: Distributed Order Matching System is initializing...")

	runner := matching.NewEngineRunner(10000)
	go runner.Start()

	server.StartGRPCServer(":50051", runner)
}
