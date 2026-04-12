package e2e_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	v1 "github.com/AleeCao/LogiTrack/gen/go/tracking/v1"
	"github.com/AleeCao/LogiTrack/pkg/config"
	"github.com/AleeCao/LogiTrack/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestE2E_FullFlow verifica el flujo completo del sistema:
// gRPC Client -> Ingestion Service -> Kafka -> Processor -> DBs (Redis, Postgres, ES)
func TestE2E_FullFlow(t *testing.T) {
	fmt.Println("========================================")
	fmt.Println(">>> E2E TEST: Starting Full Flow Test")
	fmt.Println("========================================")

	// =============================================================================
	// STEP 1: Load Configuration
	// =============================================================================
	fmt.Println("\n[STEP 1] Loading configuration...")
	env, err := config.VarConfig()
	require.NoError(t, err, "Failed to load config")
	fmt.Println("[STEP 1] Configuration loaded successfully")

	// Context with timeout for entire test
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// =============================================================================
	// STEP 2: Setup Database Connections
	// =============================================================================
	fmt.Println("\n[STEP 2] Setting up database connections...")

	// 2a. Postgres Connection
	fmt.Println("  - Connecting to Postgres...")
	pgDB, err := db.NewDBConnection(env)
	require.NoError(t, err, "Failed to connect to Postgres")
	defer pgDB.Close()
	fmt.Println("  - Postgres connected")

	// 2b. Redis Connection
	fmt.Println("  - Connecting to Redis...")
	redisClient := db.NewRedisConnection(env)
	defer redisClient.Close()
	require.NoError(t, redisClient.Ping(ctx).Err(), "Failed to ping Redis")
	fmt.Println("  - Redis connected")

	// 2c. Elasticsearch Connection
	fmt.Println("  - Connecting to Elasticsearch...")
	esClient, err := db.NewSearchEngine(env.EsAddr)
	require.NoError(t, err, "Failed to connect to Elasticsearch")
	fmt.Println("  - Elasticsearch connected")

	fmt.Println("[STEP 2] All database connections established")

	// =============================================================================
	// STEP 3: Ensure Trucks Exist in Postgres
	// =============================================================================
	fmt.Println("\n[STEP 3] Ensuring trucks exist in Postgres (for FK constraint)...")

	// Predefined trucks for the test
	truckIDs := []string{"truck-01", "truck-02", "truck-03"}
	truckModels := []string{"Tesla Semi", "Ford F-150", "Rivian EDV"}

	for i, truckID := range truckIDs {
		_, err = pgDB.ExecContext(ctx,
			"INSERT INTO trucks (truck_id, model) VALUES ($1, $2) ON CONFLICT (truck_id) DO NOTHING",
			truckID, truckModels[i])
		require.NoError(t, err, fmt.Sprintf("Failed to insert truck %s", truckID))
		fmt.Printf("  - Truck '%s' ready (model: %s)\n", truckID, truckModels[i])
	}
	fmt.Println("[STEP 3] Trucks ready in Postgres")

	// =============================================================================
	// STEP 4: Connect to Ingestion Service (gRPC Client)
	// =============================================================================
	fmt.Println("\n[STEP 4] Connecting to Ingestion Service via gRPC...")

	ingestAddr := fmt.Sprintf("localhost:%s", env.IngestionSerPort)
	conn, err := grpc.NewClient(ingestAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "Failed to connect to Ingestion Service")
	defer conn.Close()

	client := v1.NewTrackingClient(conn)

	// Open stream for each truck
	for _, truckID := range truckIDs {
		fmt.Printf("\n  >>> Opening gRPC stream for truck: %s\n", truckID)
		stream, err := client.GetLocation(ctx)
		require.NoError(t, err, fmt.Sprintf("Failed to open stream for %s", truckID))

		// =============================================================================
		// STEP 5: Send Location for Each Truck
		// =============================================================================
		fmt.Printf("\n[STEP 5] Sending location for truck: %s\n", truckID)

		// Create location with unique coordinates per truck to avoid distance check issues
		// (trucks are far apart so distance check will pass)
		locations := map[string]struct {
			lat float64
			lon float64
		}{
			"truck-01": {lat: -34.6037, lon: -58.3816}, // Buenos Aires
			"truck-02": {lat: -33.4489, lon: -70.6693}, // Santiago
			"truck-03": {lat: -23.5505, lon: -46.6333}, // São Paulo
		}

		loc := locations[truckID]

		location := &v1.Location{
			TruckId:   truckID,
			Latitude:  loc.lat,
			Longitude: loc.lon,
			Timestamp: timestamppb.Now(),
			Status:    v1.TruckStatus_ON_ITS_WAY,
		}

		fmt.Printf("  - Sending: TruckID=%s, Lat=%.4f, Lon=%.4f\n", truckID, loc.lat, loc.lon)
		err = stream.Send(location)
		require.NoError(t, err, fmt.Sprintf("Failed to send location for %s", truckID))

		// Receive ACK from server
		resp, err := stream.Recv()
		require.NoError(t, err, fmt.Sprintf("Failed to receive response for %s", truckID))

		expectedStatus := v1.DeliveryStatus_CONTINUE
		assert.Equal(t, expectedStatus, resp.Status,
			fmt.Sprintf("Expected CONTINUE status for %s, got %v", truckID, resp.Status))

		fmt.Printf("  - Received ACK: Status=%v\n", resp.Status)

		// Close stream for this truck
		stream.CloseSend()
		fmt.Printf("  >>> Stream closed for truck: %s\n", truckID)
	}
	fmt.Println("\n[STEP 5] All locations sent successfully")

	// =============================================================================
	// STEP 6: Wait for Asynchronous Processing
	// =============================================================================
	fmt.Println("\n[STEP 6] Waiting for Kafka -> Processor -> DBs (5 seconds)...")
	fmt.Println("  (This gives time for the async pipeline to complete)")
	time.Sleep(5 * time.Second)
	fmt.Println("[STEP 6] Wait complete")

	// =============================================================================
	// STEP 7: Verify Results in All Databases
	// =============================================================================
	fmt.Println("\n[STEP 7] Verifying results in all databases...")

	// Verify each truck in all databases
	for _, truckID := range truckIDs {
		fmt.Printf("\n  --- Verifying truck: %s ---\n", truckID)

		// 7a. Verify Redis (Latest Location Cache)
		fmt.Println("  [Redis] Checking latest location...")
		val, err := redisClient.HGetAll(ctx, truckID).Result()
		require.NoError(t, err, fmt.Sprintf("Redis error for %s", truckID))
		assert.NotEmpty(t, val, fmt.Sprintf("Redis should have truck %s data", truckID))
		assert.Equal(t, truckID, val["truck_ID"],
			fmt.Sprintf("Redis truck_ID mismatch for %s", truckID))
		fmt.Printf("  [Redis] ✓ Truck %s found in Redis (truck_ID=%s)\n", truckID, val["truck_ID"])

		// 7b. Verify Postgres (Location History)
		fmt.Println("  [Postgres] Checking location history...")
		var count int
		err = pgDB.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM location_history WHERE truck_id=$1", truckID).Scan(&count)
		require.NoError(t, err, fmt.Sprintf("Postgres error for %s", truckID))
		assert.GreaterOrEqual(t, count, 1,
			fmt.Sprintf("Postgres should have at least 1 record for %s", truckID))
		fmt.Printf("  [Postgres] ✓ Truck %s has %d location record(s)\n", truckID, count)

		// 7c. Verify Elasticsearch (Search Index)
		fmt.Println("  [Elasticsearch] Checking search index...")
		res, err := esClient.Get("truck_location", truckID).Do(ctx)
		require.NoError(t, err, fmt.Sprintf("Elasticsearch error for %s", truckID))
		assert.True(t, res.Found,
			fmt.Sprintf("Elasticsearch should have document for %s", truckID))
		fmt.Printf("  [Elasticsearch] ✓ Truck %s found in index\n", truckID)
	}

	fmt.Println("\n[STEP 7] All verifications passed!")

	// =============================================================================
	// TEST COMPLETE
	// =============================================================================
	fmt.Println("\n========================================")
	fmt.Println(">>> E2E TEST SUCCESSFUL!")
	fmt.Println("========================================")
	fmt.Printf("Verified %d trucks across Redis, Postgres, and Elasticsearch\n", len(truckIDs))
}
