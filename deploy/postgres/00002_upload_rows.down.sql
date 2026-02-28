TRUNCATE TABLE driver_assignment, shipments, users, trucks, drivers RESTART IDENTITY;

DROP EXTENSION IF EXISTS "pgcrypto";
