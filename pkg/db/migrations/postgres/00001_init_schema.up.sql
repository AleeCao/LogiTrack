CREATE TYPE status AS ENUM ('pending', 'in_transit', 'delivered');

CREATE TABLE IF NOT EXISTS users(
  user_id UUID PRIMARY KEY,
  username VARCHAR (50) UNIQUE NOT NULL,
  email VARCHAR (50) UNIQUE NOT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp
);

CREATE TABLE IF NOT EXISTS trucks(
  truck_id VARCHAR(20) PRIMARY KEY NOT NULL,
  model VARCHAR (30) NOT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp
);


CREATE TABLE IF NOT EXISTS shipments(
  shipment_id UUID PRIMARY KEY,
  user_id UUID REFERENCES users (user_id),
  truck_id VARCHAR REFERENCES trucks (truck_id),
  status status,
  origin_address VARCHAR(50) NOT NULL,
  destination_address VARCHAR(50) NOT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp
);


CREATE TABLE IF NOT EXISTS drivers(
  driver_id UUID PRIMARY KEY,
  driver_name VARCHAR (50) NOT NULL,
  phone_number VARCHAR (15) UNIQUE NOT NULL,
  license_number VARCHAR (15) UNIQUE NOT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp
);


CREATE TABLE IF NOT EXISTS driver_assignment(
  assignment_id UUID PRIMARY KEY,
  driver_id UUID REFERENCES drivers (driver_id),
  truck_id VARCHAR(20) REFERENCES trucks (truck_id),
  start_time timestamp NOT NULL,
  end_time timestamp NOT NULL
);
