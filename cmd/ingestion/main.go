package main

import (
	"fmt"
	"log"
	"net"

	v1 "github.com/AleeCao/LogiTrack/gen/go/tracking/v1"
	"github.com/AleeCao/LogiTrack/internal/ingestion/adapters"
	"github.com/AleeCao/LogiTrack/internal/ingestion/services"
	"github.com/AleeCao/LogiTrack/pkg/config"
	"google.golang.org/grpc"
)

func main() {
	env, err := config.VarConfig()
	if err != nil {
		log.Fatal(err)
	}
	broker := adapters.NewProducer(fmt.Sprintf("%s:%s", env.KafkaHost, env.KafkaPort), env.KafkaTopic)
	ingSvc := services.NewIngestionService(broker)
	grpcAdp := adapters.NewGrpcAdapter(ingSvc)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", env.IngestionSerHost, env.IngestionSerPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	v1.RegisterTrackingServer(grpcServer, grpcAdp)
	log.Printf("Listen and serve on port: %s", env.IngestionSerPort)
	grpcServer.Serve(lis)
}
