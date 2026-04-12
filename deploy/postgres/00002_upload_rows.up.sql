CREATE TABLE IF NOT EXISTS location_history (
       id BIGSERIAL PRIMARY KEY,
       truck_id VARCHAR(20) NOT NULL REFERENCES trucks(truck_id),
       latitude DOUBLE PRECISION NOT NULL,
       longitude DOUBLE PRECISION NOT NULL,
       truck_status VARCHAR(20),
       timestamp TIMESTAMP WITH TIME ZONE NOT NULL
   );

   CREATE INDEX idx_location_history_truck_id ON location_history(truck_id);
   CREATE INDEX idx_location_history_timestamp ON location_history(timestamp);
