package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/AleeCao/LogiTrack/internal/domain"
	"github.com/AleeCao/LogiTrack/internal/processor/ports"
)

type BatcherService struct {
	StorageRepo ports.StorageRepository
	CacheRepo   ports.CacheRepository
}

func NewBatcherService(str ports.StorageRepository, ca ports.CacheRepository) ports.BatchService {
	return &BatcherService{StorageRepo: str, CacheRepo: ca}
}

func (s *BatcherService) DbBatcher(ctx context.Context, locations <-chan *domain.Location, wg *sync.WaitGroup) {
	defer wg.Done()
	const timeOut = 3 * time.Second
	timer := time.NewTimer(timeOut)
	defer timer.Stop()

	var buffer []*domain.Location

	for {
		select {
		case op := <-locations:
			buffer = append(buffer, op)
			if len(buffer) >= 100 {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
					s.StorageRepo.InsertLocationRecord(ctx, &buffer)
					timer.Reset(timeOut)
				}
			}
		case <-timer.C:
			if len(buffer) == 0 {
				timer.Reset(timeOut)
			} else {
				s.StorageRepo.InsertLocationRecord(ctx, &buffer)
				timer.Reset(timeOut)
			}
		case <-ctx.Done():
			if len(buffer) != 0 {
				s.StorageRepo.InsertLocationRecord(ctx, &buffer)
			}
			log.Println("Batcher shutting down")
			return
		}
	}
}

func (s *BatcherService) CacheBatcher(ctx context.Context, locations <-chan *domain.Location, wg *sync.WaitGroup) {
	defer wg.Done()
	const timeOut = 3 * time.Second
	timer := time.NewTimer(timeOut)
	defer timer.Stop()

	var buffer []*domain.Location

	for {
		select {
		case op := <-locations:
			buffer = append(buffer, op)
			if len(buffer) >= 100 {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
					s.CacheRepo.SendBatch(ctx, &buffer)
					timer.Reset(timeOut)
				}
			}
		case <-timer.C:
			if len(buffer) == 0 {
				timer.Reset(timeOut)
			} else {
				s.CacheRepo.SendBatch(ctx, &buffer)
				timer.Reset(timeOut)
			}
		case <-ctx.Done():
			if len(buffer) != 0 {
				s.CacheRepo.SendBatch(ctx, &buffer)
			}
			log.Println("Batcher shutting down")
			return
		}
	}
}
