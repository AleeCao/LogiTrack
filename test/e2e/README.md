# Test E2E - LogiTrack

## Descripción

Este test verifica el flujo completo del sistema LogiTrack, validando que todos los componentes funcionen correctamente juntos:

```
gRPC Client → Ingestion Service → Kafka → Processor → DBs (Redis, Postgres, Elasticsearch)
```

## Qué hace el test

1. **Conexiones**: Establece conexiones a Postgres, Redis y Elasticsearch
2. **Prepara datos**: Asegura que los trucks existan en Postgres (tabla `trucks`) para cumplir con la restricción FK
3. **Envía ubicaciones**: Simula 3 trucks enviando su ubicación via gRPC al Ingestion Service
4. **Procesa asincrónicamente**: Espera a que el mensaje fluya por Kafka → Processor → Bases de datos
5. **Verifica**: Confirma que cada truck esté persistido en las 3 bases de datos

## Trucks utilizados

| Truck ID  | Modelo       | Ubicación simulada |
|-----------|--------------|---------------------|
| truck-01  | Tesla Semi   | Buenos Aires        |
| truck-02  | Ford F-150   | Santiago            |
| truck-03  | Rivian EDV   | São Paulo           |

## Requisitos previos

Para que el test funcione, deben estar corriendo:

- **Docker Compose**: Postgres, Redis, Elasticsearch, Kafka
- **Ingestion Service** (binario Go): Servicio gRPC en puerto 8080
- **Processor Service** (binario Go): Consumidor de Kafka

## Logs esperados del test (Éxito)

```
========================================
>>> E2E TEST: Starting Full Flow Test
========================================

[STEP 1] Loading configuration...
[STEP 1] Configuration loaded successfully

[STEP 2] Setting up database connections...
  - Connecting to Postgres...
  - Postgres connected
  - Connecting to Redis...
  - Redis connected
  - Connecting to Elasticsearch...
  - Elasticsearch connected
[STEP 2] All database connections established

[STEP 3] Ensuring trucks exist in Postgres (for FK constraint)...
  - Truck 'truck-01' ready (model: Tesla Semi)
  - Truck 'truck-02' ready (model: Ford F-150)
  - Truck 'truck-03' ready (model: Rivian EDV)
[STEP 3] Trucks ready in Postgres

[STEP 4] Connecting to Ingestion Service via gRPC...

  >>> Opening gRPC stream for truck: truck-01

[STEP 5] Sending location for truck: truck-01
  - Sending: TruckID=truck-01, Lat=-34.6037, Lon=-58.3816
  - Received ACK: Status=CONTINUE
  >>> Stream closed for truck: truck-01

  >>> Opening gRPC stream for truck: truck-02

[STEP 5] Sending location for truck: truck-02
  - Sending: TruckID=truck-02, Lat=-33.4489, Lon=-70.6693
  - Received ACK: Status=CONTINUE
  >>> Stream closed for truck: truck-02

  >>> Opening gRPC stream for truck: truck-03

[STEP 5] Sending location for truck: truck-03
  - Sending: TruckID=truck-03, Lat=-23.5505, Lon=-46.6333
  - Received ACK: Status=CONTINUE
  >>> Stream closed for truck: truck-03

[STEP 5] All locations sent successfully

[STEP 6] Waiting for Kafka -> Processor -> DBs (5 seconds)
  (This gives time for the async pipeline to complete)
[STEP 6] Wait complete

[STEP 7] Verifying results in all databases...

  --- Verifying truck: truck-01 ---
  [Redis] Checking latest location...
  [Redis] ✓ Truck truck-01 found in Redis (truck_ID=truck-01)
  [Postgres] Checking location history...
  [Postgres] ✓ Truck truck-01 has 1 location record(s)
  [Elasticsearch] Checking search index...
  [Elasticsearch] ✓ Truck truck-01 found in index

  --- Verifying truck: truck-02 ---
  [Redis] Checking latest location...
  [Redis] ✓ Truck truck-02 found in Redis (truck_ID=truck-02)
  [Postgres] Checking location history...
  [Postgres] ✓ Truck truck-02 has 1 location record(s)
  [Elasticsearch] Checking search index...
  [Elasticsearch] ✓ Truck truck-02 found in index

  --- Verifying truck: truck-03 ---
  [Redis] Checking latest location...
  [Redis] ✓ Truck truck-03 found in Redis (truck_ID=truck-03)
  [Postgres] Checking location history...
  [Postgres] ✓ Truck truck-03 has 1 location record(s)
  [Elasticsearch] Checking search index...
  [Elasticsearch] ✓ Truck truck-03 found in index

[STEP 7] All verifications passed!

========================================
>>> E2E TEST SUCCESSFUL!
========================================
Verified 3 trucks across Redis, Postgres, and Elasticsearch
```

## Logs del Ingestion Service (Éxito)

El Ingestion Service debe mostrar logs similares a:

```
2026/04/12 10:00:00 Listen and serve on port: 8080
2026/04/12 10:00:05 Location published: &{TruckID:truck-01 Latitude:-34.6037 Longitude:-58.3816 ...}
2026/04/12 10:00:06 Location published: &{TruckID:truck-02 Latitude:-33.4489 Longitude:-70.6693 ...}
2026/04/12 10:00:07 Location published: &{TruckID:truck-03 Latitude:-23.5505 Longitude:-46.6333 ...}
```

**Qué significa cada línea**:
- `Listen and serve on port: 8080` → El servicio gRPC está escuchando
- `Location published: ...` → El mensaje fue enviado a Kafka exitosamente

## Logs del Processor Service (Éxito)

El Processor debe mostrar:

```
Consumer initialized with topic: TruckLocation
start consuming
Waiting for message...
Received message: {"TruckID":"truck-01",...}
Processing location for truck: truck-01
Distance check passed, updating storage...
Error: <nil> (o nada si todo sale bien)
```

**Qué significa cada línea**:

| Log | Significado |
|-----|-------------|
| `Consumer initialized with topic: TruckLocation` | El consumidor se conectó al topic de Kafka |
| `start consuming` | El consumidor comenzó a escuchar mensajes |
| `Waiting for message...` | Esperando mensajes del broker |
| `Received message: ...` | Mensaje recibido de Kafka |
| `Processing location for truck: truck-01` | Processor inició procesamiento |
| `Distance check passed, updating storage...` | La ubicación pasó la validación de distancia (>500m) y se guardará |
| `Distance check failed (too close)` | La ubicación está a menos de 500m de la anterior, no se guarda en Postgres/Redis pero sí en Elasticsearch |

## Posibles errores y qué significan

### Error en el test: `Redis should have truck data`

**Causa probable**: El Processor no guardó la ubicación en Redis

**Qué verificar**:
1. ¿El Processor está corriendo?
2. ¿Los logs del Processor muestran `Distance check passed, updating storage...`?
3. ¿Hay error en `SetLocationRecord` en los logs?

### Error en el test: `Postgres should have at least 1 record`

**Causa probable**: No se insertó en la tabla `location_history`

**Qué verificar**:
1. ¿La tabla `trucks` tiene el truck dado de alta? (FK constraint)
2. ¿Los logs del Processor muestran `Error creating location record`?
3. ¿Hay constraint de FK y está fallando silenciosamente?

### Error en el test: `Elasticsearch should have the document indexed`

**Causa probable**: El Processor no actualizó el índice de Elasticsearch

**Qué verificar**:
1. ¿Los logs del Processor muestran `Error updating search repo`?
2. ¿Elasticsearch está corriendo y accessible?

### Error: `failed consuming from broker: context canceled`

**Causa probable**: El Processor recibió una señal de cancelación (Ctrl+C o shutdown)

**Qué verificar**:
1. ¿El proceso del Processor sigue corriendo?
2. ¿Se cortó manualmente el proceso?

### Error en Ingestion Service: `Failed to connect to Ingestion Service`

**Causa probable**: El servicio de ingestión no está corriendo

**Qué verificar**:
1. ¿El binary del Ingestion Service está corriendo?
2. ¿El puerto 8080 está disponible?

## Cómo ejecutar el test

```bash
# 1. Levantar servicios de Docker
docker-compose up -d

# 2. Correr el Ingestion Service (en otra terminal)
go run cmd/ingestion/main.go

# 3. Correr el Processor Service (en otra terminal)
go run cmd/processor/main.go

# 4. Ejecutar el test
go test ./test/e2e/... -v
```

## Notas adicionales

- El test espera **5 segundos** para que el procesamiento asincrónico complete. Si los servicios están lentos, este tiempo puede no ser suficiente.
- Los trucks usan coordenadas geográficas suficientemente separadas para evitar el filtro de distancia (500m).
- Los datos no se limpian entre ejecuciones, por lo que runs consecutivos acumulan registros en Postgres.