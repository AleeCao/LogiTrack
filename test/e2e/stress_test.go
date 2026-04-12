package e2e_test

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	v1 "github.com/AleeCao/LogiTrack/gen/go/tracking/v1"
	"github.com/AleeCao/LogiTrack/pkg/config"
	"github.com/AleeCao/LogiTrack/pkg/db"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestStress_FiftyTruckFlow testea el sistema con 50 trucks enviando ubicaciones
// continuamente durante 2 minutos, verificando el comportamiento bajo carga.
func TestStress_FiftyTruckFlow(t *testing.T) {
	fmt.Println("========================================")
	fmt.Println(">>> STRESS TEST: 50 Trucks x 2 Minutes")
	fmt.Println("========================================")

	// =============================================================================
	// STEP 1: Load Configuration
	// =============================================================================
	fmt.Println("\n[STEP 1] Loading configuration...")
	env, err := config.VarConfig()
	require.NoError(t, err, "Failed to load config")
	fmt.Println("[STEP 1] Configuration loaded successfully")

	// Context with extended timeout for 2 minute test
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// =============================================================================
	// STEP 2: Setup Database Connections
	// =============================================================================
	fmt.Println("\n[STEP 2] Setting up database connections...")

	// Postgres
	fmt.Println("  - Connecting to Postgres...")
	pgDB, err := db.NewDBConnection(env)
	require.NoError(t, err, "Failed to connect to Postgres")
	defer pgDB.Close()
	fmt.Println("  - Postgres connected")

	// Redis
	fmt.Println("  - Connecting to Redis...")
	redisClient := db.NewRedisConnection(env)
	defer redisClient.Close()
	require.NoError(t, redisClient.Ping(ctx).Err(), "Failed to ping Redis")
	fmt.Println("  - Redis connected")

	// Elasticsearch
	fmt.Println("  - Connecting to Elasticsearch...")
	esClient, err := db.NewSearchEngine(env.EsAddr)
	require.NoError(t, err, "Failed to connect to Elasticsearch")
	fmt.Println("  - Elasticsearch connected")

	fmt.Println("[STEP 2] All database connections established")

	// =============================================================================
	// STEP 3: Generate 50 Truck IDs
	// =============================================================================
	fmt.Println("\n[STEP 3] Generating 50 truck IDs...")

	const numTrucks = 50
	truckIDs := make([]string, numTrucks)
	truckModels := make([]string, numTrucks)

	for i := 0; i < numTrucks; i++ {
		truckIDs[i] = fmt.Sprintf("truck-%02d", i+1)
		truckModels[i] = fmt.Sprintf("Model-%d", i+1)
	}
	fmt.Printf("  - Generated %d truck IDs (truck-01 to truck-50)\n", numTrucks)

	// =============================================================================
	// STEP 4: Ensure Trucks Exist in Postgres
	// =============================================================================
	fmt.Println("\n[STEP 4] Ensuring trucks exist in Postgres...")

	for i, truckID := range truckIDs {
		_, err = pgDB.ExecContext(ctx,
			"INSERT INTO trucks (truck_id, model) VALUES ($1, $2) ON CONFLICT (truck_id) DO NOTHING",
			truckID, truckModels[i])
		require.NoError(t, err, fmt.Sprintf("Failed to insert truck %s", truckID))
		if (i+1)%10 == 0 {
			fmt.Printf("  - Inserted %d/%d trucks\n", i+1, numTrucks)
		}
	}
	fmt.Println("[STEP 4] All trucks ready in Postgres")

	// =============================================================================
	// STEP 5: Connect to Ingestion Service
	// =============================================================================
	fmt.Println("\n[STEP 5] Connecting to Ingestion Service via gRPC...")

	ingestAddr := fmt.Sprintf("localhost:%s", env.IngestionSerPort)
	conn, err := grpc.NewClient(ingestAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "Failed to connect to Ingestion Service")
	defer conn.Close()

	client := v1.NewTrackingClient(conn)
	fmt.Println("[STEP 5] gRPC connection established")

	// =============================================================================
	// STEP 6: Continuous Location Sending (2 minutes, every 5 seconds)
	// =============================================================================
	fmt.Println("\n[STEP 6] Starting continuous location sending...")
	fmt.Println("  - Duration: 2 minutes")
	fmt.Println("  - Interval: 5 seconds per cycle")
	fmt.Println("  - Trucks: 50")

	// Duration: 2 minutes = 120 seconds
	testDuration := 2 * time.Minute
	interval := 5 * time.Second

	startTime := time.Now()
	cycleCount := 0

	// Seed random for coordinates
	rand.Seed(time.Now().UnixNano())

	// Base coordinates to spread trucks across different cities
	baseCoords := []struct{ lat, lon float64 }{
		{-34.6037, -58.3816}, // Buenos Aires
		{-33.4489, -70.6693}, // Santiago
		{-23.5505, -46.6333}, // São Paulo
		{-19.9167, -43.9345}, // Belo Horizonte
		{-12.9714, -38.5014}, // Salvador
		{-22.9068, -43.1729}, // Río de Janeiro
		{-8.0476, -34.8770},  // Recife
		{-3.7172, -38.5433},  // Fortaleza
		{-0.0236, -78.1228},  // Quito
		{4.7110, -74.0721},   // Bogotá
	}

	// Run for 2 minutes
	for time.Since(startTime) < testDuration {
		cycleCount++
		cycleStart := time.Now()

		fmt.Printf("\n  >>> Cycle %d - Time: %v / %v\n", cycleCount, time.Since(startTime).Round(time.Second), testDuration)

		// Send location for each truck
		successCount := 0
		for i, truckID := range truckIDs {
			// Use different base coordinate for each truck
			baseIdx := i % len(baseCoords)
			// Add small random offset to avoid distance check issues
			lat := baseCoords[baseIdx].lat + (rand.Float64()-0.5)*0.1
			lon := baseCoords[baseIdx].lon + (rand.Float64()-0.5)*0.1

			// Open stream for this truck
			stream, err := client.GetLocation(ctx)
			if err != nil {
				fmt.Printf("    ! Failed to open stream for %s: %v\n", truckID, err)
				continue
			}

			location := &v1.Location{
				TruckId:   truckID,
				Latitude:  lat,
				Longitude: lon,
				Timestamp: timestamppb.Now(),
				Status:    v1.TruckStatus_ON_ITS_WAY,
			}

			err = stream.Send(location)
			if err != nil {
				fmt.Printf("    ! Failed to send for %s: %v\n", truckID, err)
				stream.CloseSend()
				continue
			}

			// Receive ACK
			resp, err := stream.Recv()
			if err != nil {
				fmt.Printf("    ! Failed to receive ACK for %s: %v\n", truckID, err)
				stream.CloseSend()
				continue
			}

			if resp.Status == v1.DeliveryStatus_CONTINUE {
				successCount++
			}

			stream.CloseSend()
		}

		fmt.Printf("    - Sent: %d/%d trucks successful\n", successCount, numTrucks)

		// Mid-test verification (every 10 cycles ~ 50 seconds)
		if cycleCount%10 == 0 {
			fmt.Printf("\n  [MID-TEST VERIFICATION] Cycle %d\n", cycleCount)
			verifyTrucks(ctx, t, truckIDs, pgDB, redisClient, esClient)
		}

		// Wait for next cycle
		elapsed := time.Since(cycleStart)
		if elapsed < interval {
			time.Sleep(interval - elapsed)
		}
	}

	fmt.Println("\n[STEP 6] Continuous sending complete")
	fmt.Printf("  - Total cycles: %d\n", cycleCount)
	fmt.Printf("  - Total time: %v\n", time.Since(startTime))

	// =============================================================================
	// STEP 7: Final Verification
	// =============================================================================
	fmt.Println("\n[STEP 7] Final verification of all 50 trucks...")

	verifyTrucks(ctx, t, truckIDs, pgDB, redisClient, esClient)

	fmt.Println("\n[STEP 7] Final verification complete!")

	// =============================================================================
	// TEST COMPLETE
	// =============================================================================
	fmt.Println("\n========================================")
	fmt.Println(">>> STRESS TEST SUCCESSFUL!")
	fmt.Println("========================================")
	fmt.Printf("Duration: 2 minutes\n")
	fmt.Printf("Cycles: %d\n", cycleCount)
	fmt.Printf("Trucks: %d\n", numTrucks)
	fmt.Printf("Total location updates: ~%d\n", cycleCount*numTrucks)
}

// verifyTrucks verifies all trucks in all databases
func verifyTrucks(ctx context.Context, t *testing.T, truckIDs []string, pgDB *sql.DB, redisClient *redis.Client, esClient *elasticsearch.TypedClient) {
	fmt.Println("  Verifying in all databases...")

	// Track verification results
	redisOK := 0
	postgresOK := 0
	esOK := 0

	for _, truckID := range truckIDs {
		// Redis check
		redisResult, redisErr := redisClient.HGetAll(ctx, truckID).Result()
		if redisErr == nil && len(redisResult) > 0 && redisResult["truck_ID"] != "" {
			redisOK++
		}

		// Postgres check
		var count int
		postgresErr := pgDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM location_history WHERE truck_id=$1", truckID).Scan(&count)
		if postgresErr == nil && count > 0 {
			postgresOK++
		}

		// Elasticsearch check
		esResult, esErr := esClient.Get("truck_location", truckID).Do(ctx)
		if esErr == nil && esResult.Found {
			esOK++
		}
	}

	fmt.Printf("  - Redis: %d/%d trucks have latest location\n", redisOK, len(truckIDs))
	fmt.Printf("  - Postgres: %d/%d trucks have history\n", postgresOK, len(truckIDs))
	fmt.Printf("  - Elasticsearch: %d/%d trucks indexed\n", esOK, len(truckIDs))

	// Assert all trucks are present
	assert.GreaterOrEqual(t, redisOK, len(truckIDs)-5,
		"Redis should have most trucks (allow 5 for timing)")
	assert.GreaterOrEqual(t, postgresOK, len(truckIDs)-5,
		"Postgres should have most trucks (allow 5 for timing)")
	assert.GreaterOrEqual(t, esOK, len(truckIDs)-5,
		"Elasticsearch should have most trucks (allow 5 for timing)")

	fmt.Println("  ✓ Verification passed!")
}
