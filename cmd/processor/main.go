package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/AleeCao/LogiTrack/internal/processor/adapters"
	"github.com/AleeCao/LogiTrack/internal/processor/services"
	"github.com/AleeCao/LogiTrack/pkg/config"
	"github.com/AleeCao/LogiTrack/pkg/db"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	envVar, err := config.VarConfig()
	if err != nil {
		log.Fatal("%v", err)
	}

	dbPro, err := db.NewDBConnection(envVar)
	if err != nil {
		log.Fatal("%v", err)
	}

	cacheClient := db.NewRedisConnection(envVar)

	searchEngine, err := db.NewSearchEngine(envVar.EsAddr)
	if err != nil {
		log.Fatal("%v", err)
	}

	cacheRep := adapters.NewCacheRepository(cacheClient)

	storageRep := adapters.NewStorageRepository(dbPro)

	searchRep := adapters.NewSearchRepo(searchEngine)

	processorSvc := services.NewProcessorService(storageRep, cacheRep, searchRep)

	consumerAddr := fmt.Sprintf("%s:%s", envVar.KafkaHost, envVar.KafkaPort)
	consumerWorker := adapters.NewConsumer(consumerAddr, envVar.KafkaTopic, processorSvc)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("shutting down")
		cancel()
	}()

	if err := consumerWorker.StartConsume(ctx); err != nil {
		log.Fatalf("Consumer failed: %v", err)
	}
}
