package adapters_test

import (
	"context"
	"testing"
	"time"

	"github.com/AleeCao/LogiTrack/internal/domain"
	"github.com/AleeCao/LogiTrack/internal/processor/adapters"
	"github.com/AleeCao/LogiTrack/pkg/config"
	"github.com/AleeCao/LogiTrack/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAdaptersIntegration ensures that our adapters correctly interact with
// real infrastructure (Postgres and Redis).
// Note: This test requires the Docker-compose stack to be running.
func TestAdapters_Integration(t *testing.T) {
	// Initialize configuration
	env, err := config.VarConfig()
	require.NoError(t, err, "Failed to load environment configuration")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize PostgreSQL connection
	sqlDB, err := db.NewDBConnection(env)
	require.NoError(t, err, "Failed to connect to PostgreSQL")
	t.Cleanup(func() { sqlDB.Close() })

	// Initialize Redis connection
	redisClient := db.NewRedisConnection(env)
	t.Cleanup(func() { redisClient.Close() })

	// Ensure Redis is reachable before proceeding
	err = redisClient.Ping(ctx).Err()
	require.NoError(t, err, "Redis is not reachable; ensure docker-compose is up")

	// Initialize the adapters under test
	storageRepo := adapters.NewStorageRepository(sqlDB)
	cacheRepo := adapters.NewCacheRepository(redisClient)

	// Prepare shared test data
	testTruckID := "integration-test-truck-001"
	testLocation := &domain.Location{
		TruckID:     testTruckID,
		Latitude:    -34.6037, // Buenos Aires coordinates
		Longitude:   -58.3816,
		UpdatedAt:   time.Now().Truncate(time.Second).UTC(),
		TruckStatus: "ON_ITS_WAY",
	}

	// 1. Cache Repository (Redis) Sub-tests
	t.Run("CacheRepository: Save and Retrieve", func(t *testing.T) {
		// Cleanup existing keys to ensure a clean state
		redisClient.Del(ctx, testTruckID)

		// Persist the location record to cache
		err := cacheRepo.SetLocationRecord(ctx, testLocation)
		assert.NoError(t, err, "Should successfully save location to Redis")

		// Retrieve and verify the record
		retrieved, err := cacheRepo.GetLocationRecord(ctx, testTruckID)
		assert.NoError(t, err, "Should successfully retrieve record from Redis")
		require.NotNil(t, retrieved)

		assert.Equal(t, testLocation.TruckID, retrieved.TruckID)
		assert.InDelta(t, testLocation.Latitude, retrieved.Latitude, 0.0001)
		assert.InDelta(t, testLocation.Longitude, retrieved.Longitude, 0.0001)
		assert.Equal(t, testLocation.TruckStatus, retrieved.TruckStatus)

		// Note: We use InDelta for coordinates due to potential floating point precision loss during serialization
	})

	t.Run("CacheRepository: Handle Missing Key", func(t *testing.T) {
		const nonExistentID = "truck-does-not-exist"

		redisClient.Del(ctx, nonExistentID)

		retrieved, err := cacheRepo.GetLocationRecord(ctx, nonExistentID)
		assert.NoError(t, err, "Fetching a missing key should not return an error")
		assert.Nil(t, retrieved, "Expected nil for a non-existent truck ID")
	})

	// 2. Storage Repository (Postgres) Sub-tests
	t.Run("StorageRepository: Record History", func(t *testing.T) {
		// Ensure the truck exists in the master table due to Foreign Key constraints
		const upsertTruckQuery = `
			INSERT INTO trucks (truck_id, model) 
			VALUES ($1, $2) 
			ON CONFLICT (truck_id) DO NOTHING`

		_, err := sqlDB.ExecContext(ctx, upsertTruckQuery, testTruckID, "Testing-Model-X")
		require.NoError(t, err, "Setup: Failed to ensure truck exists in the database")

		// Create a historical location record
		err = storageRepo.CreateLocationRecord(ctx, testLocation)
		assert.NoError(t, err, "Should successfully insert historical record in Postgres")

		// Validate persistence
		var exists bool
		const verifyQuery = "SELECT EXISTS(SELECT 1 FROM location_history WHERE truck_id=$1 AND timestamp=$2)"
		err = sqlDB.QueryRowContext(ctx, verifyQuery, testTruckID, testLocation.UpdatedAt).Scan(&exists)

		assert.NoError(t, err, "Failed to query the database for verification")
		assert.True(t, exists, "The location record was not found in the database")
	})
}
