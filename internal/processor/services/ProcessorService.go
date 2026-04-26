// Package services implements the business logic for real-time location processing,
// including distance calculation rules (500m threshold) and orchestration
// between cache, persistent storage, and search engine layers.
package services

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/AleeCao/LogiTrack/internal/domain"
	"github.com/AleeCao/LogiTrack/internal/processor/ports"
)

type ProcessorService struct {
	StorageRepo ports.StorageRepository
	CacheRepo   ports.CacheRepository
	SearchRepo  ports.SearchRepository
}

func NewProcessorService(str ports.StorageRepository, ca ports.CacheRepository, srch ports.SearchRepository) ports.ProcessService {
	return &ProcessorService{StorageRepo: str, CacheRepo: ca, SearchRepo: srch}
}

func (s *ProcessorService) ProcessLocation(ctx context.Context, locations <-chan domain.Location, wg *sync.WaitGroup) error {
	defer wg.Done()
	for lcn := range locations {
		fmt.Printf("Processing location for truck: %s\n", lcn.TruckID)
		lastLcn, err := s.CacheRepo.GetLocationRecord(ctx, lcn.TruckID)
		if err != nil {
			fmt.Printf("Error getting location record: %v\n", err)
			return err
		}
		if s.checkLastUpdate(lastLcn, &lcn) {
			fmt.Println("Last Update check passed")
			if s.checkDistance(&lcn, lastLcn) {
				fmt.Println("Distance check passed")
				if err := s.CacheRepo.SetLocationRecord(ctx, &lcn); err != nil {
					fmt.Printf("Error setting location record: %v\n", err)
					return err
				}
				if err := s.StorageRepo.CreateLocationRecord(ctx, &lcn); err != nil {
					fmt.Printf("Error creating location record: %v\n", err)
					return err
				}
			} else {
				fmt.Println("Distance check failed")
			}
		} else {
			fmt.Println("Last Update check failed")
		}
		if err := s.SearchRepo.UpdateTruckLocation(ctx, &lcn); err != nil {
			fmt.Printf("Error updating search repo: %v\n", err)
			return err
		}
	}
	fmt.Println("Worker shutting down")
	return nil
}

func (s *ProcessorService) checkLastUpdate(lcn1 *domain.Location, lcn2 *domain.Location) bool {
	if lcn2 == nil {
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
