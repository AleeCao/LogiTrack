// Package adapters provides implementation-specific code to interact with
// external services such as PostgreSQL for persistent storage, Redis for
// caching latest locations, and Elasticsearch for geospatial searching.
package adapters

import (
	"context"
	"fmt"

	"github.com/AleeCao/LogiTrack/internal/domain"
	processorDomain "github.com/AleeCao/LogiTrack/internal/processor/domain"
	"github.com/AleeCao/LogiTrack/internal/processor/ports"
	"github.com/elastic/go-elasticsearch/v9"
)

type SearchRepo struct {
	es *elasticsearch.TypedClient
}

func NewSearchRepo(es *elasticsearch.TypedClient) ports.SearchRepository {
	return &SearchRepo{es}
}

func (s *SearchRepo) UpdateTruckLocation(ctx context.Context, lcn *domain.Location) error {
	geoLcn := processorDomain.EsLocation{
		TruckID: lcn.TruckID,
		Location: processorDomain.GeoPoint{
			Longitude: lcn.Longitude,
			Latitude:  lcn.Latitude,
		},
		UpdatedAt:   lcn.UpdatedAt,
		TruckStatus: lcn.TruckStatus,
	}

	_, err := s.es.Index(
		"truck_location").Id(geoLcn.TruckID).Request(geoLcn).Do(ctx)
	if err != nil {
		return fmt.Errorf("error indexing document: %w", err)
	}

	return nil
}
