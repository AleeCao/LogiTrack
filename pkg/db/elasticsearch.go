package db

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v9"
)

type SearchEngine struct {
	es *elasticsearch.TypedClient
}

func NewSearchEngine(addrs string) (*elasticsearch.TypedClient, error) {
	es, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{addrs},
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to elasticsearch: %w", err)
	}

	return es, nil
}
