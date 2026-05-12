// Package services implements the business logic for real-time location processing,
// including distance calculation rules (500m threshold) and orchestration
// between cache, persistent storage, and search engine layers.
package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/AleeCao/LogiTrack/internal/domain"
	"github.com/AleeCao/LogiTrack/internal/processor/ports"
)

type ProcessorService struct {
	CacheRepo  ports.CacheRepository
	SearchRepo ports.SearchRepository
}

func NewProcessorService(ca ports.CacheRepository, srch ports.SearchRepository) ports.ProcessService {
	return &ProcessorService{CacheRepo: ca, SearchRepo: srch}
}

func (s *ProcessorService) ProcessLocationRecord(ctx context.Context, locations <-chan *domain.Location, cacheChan chan *domain.Location, storageChan chan *domain.Location, wg *sync.WaitGroup) error {
	defer wg.Done()

	for currentLcn := range locations {
		fmt.Printf("Processing location for truck: %s\n", currentLcn.TruckID)
		lastLcn, err := s.CacheRepo.GetLocationRecord(ctx, currentLcn.TruckID)
		if err != nil {
			fmt.Printf("Error getting location record: %v\n", err)
			return err
		}
		if s.checkLastUpdate(lastLcn, currentLcn) {
			fmt.Println("Last Update check passed")
			if s.checkDistance(currentLcn, lastLcn) {
				fmt.Println("Distance check passed")
				select {
				case cacheChan <- currentLcn:
				case <-ctx.Done():
					log.Println("Context cancelled while publishing on cache channel")
					return ctx.Err()
				}

				select {
				case storageChan <- currentLcn:
				case <-ctx.Done():
					log.Println("Context cancelled while publishing on strage channel")
					return ctx.Err()
				}
			} else {
				fmt.Println("Distance check failed")
			}
		} else {
			fmt.Println("Last Update check failed")
		}
	}
	fmt.Println("Worker shutting down")
	return nil
}

func (s *ProcessorService) ProcessLocationSearch(ctx context.Context, locations <-chan *domain.Location, wg *sync.WaitGroup) error {
	defer wg.Done()

	for currentLcn := range locations {
		if err := s.SearchRepo.UpdateTruckLocation(ctx, currentLcn); err != nil {
			fmt.Printf("Error updating search repo: %v\n", err)
			return err
		}
	}
	log.Println("Worker shutting down")
	return nil
}

func (s *ProcessorService) checkLastUpdate(lcn1 *domain.Location, lcn2 *domain.Location) bool {
	if lcn1 == nil {
		return true
	}
	return lcn2.UpdatedAt.After(lcn1.UpdatedAt)
}

func (s *ProcessorService) checkDistance(lcn1 *domain.Location, lcn2 *domain.Location) bool {
	if lcn2 == nil {
		return true
	}

	const earthRadius = 6371000

	lat1 := lcn1.Latitude * math.Pi / 180
	lon1 := lcn1.Longitude * math.Pi / 180
	lat2 := lcn2.Latitude * math.Pi / 180
	lon2 := lcn2.Longitude * math.Pi / 180

	dLat := lat2 - lat1
	dLon := lon2 - lon1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := earthRadius * c

	return distance > 500
}
