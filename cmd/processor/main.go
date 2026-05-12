package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/AleeCao/LogiTrack/internal/domain"
	"github.com/AleeCao/LogiTrack/internal/processor/adapters"
	"github.com/AleeCao/LogiTrack/internal/processor/services"
	"github.com/AleeCao/LogiTrack/pkg/config"
	"github.com/AleeCao/LogiTrack/pkg/db"
)

func main() {
	workersChan, esChan, cacheChan, storageChan := make(chan *domain.Location, 500), make(chan *domain.Location, 500), make(chan *domain.Location, 500), make(chan *domain.Location, 500)
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	envVar, err := config.VarConfig()
	if err != nil {
		log.Fatalf("%v", err)
	}

	dbPro, err := db.NewDBConnection(ctx, envVar)
	if err != nil {
		log.Fatalf("%v", err)
	}

	cacheClient := db.NewRedisConnection(envVar)

	searchEngine, err := db.NewSearchEngine(envVar.EsAddr)
	if err != nil {
		log.Fatalf("%v", err)
	}

	cacheRep := adapters.NewCacheRepository(cacheClient)

	storageRep := adapters.NewStorageRepository(dbPro)

	searchRep := adapters.NewSearchRepo(searchEngine)

	processorSvc := services.NewProcessorService(cacheRep, searchRep)

	batcherSvc := services.NewBatcherService(storageRep, cacheRep)

	consumerAddr := fmt.Sprintf("%s:%s", envVar.KafkaHost, envVar.KafkaPort)
	consumerWorker := adapters.NewConsumer(consumerAddr, envVar.KafkaTopic, processorSvc)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("shutting down")
		cancel()
		close(esChan)
		close(workersChan)
		close(cacheChan)
		close(storageChan)
		wg.Wait()
		fmt.Println("All workers have processed thier messages.")
	}()

	for i := 0; i < (runtime.NumCPU() / 4); i++ {
		wg.Add(2)
		go processorSvc.ProcessLocationRecord(ctx, workersChan, cacheChan, storageChan, &wg)
		go processorSvc.ProcessLocationSearch(ctx, esChan, &wg)
	}

	for i := 0; i < 2; i++ {
		wg.Add(2)
		go batcherSvc.CacheBatcher(ctx, cacheChan, &wg)
		go batcherSvc.DbBatcher(ctx, storageChan, &wg)
	}

	if err := consumerWorker.StartConsume(ctx, workersChan, esChan); err != nil {
		log.Fatalf("Consumer failed: %v", err)
	}
}
