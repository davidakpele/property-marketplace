# Property Marketplace API

A property listing REST API built with Go, PostgreSQL + PostGIS, and Redis. It supports full CRUD for listings, geospatial radius search, Redis-backed caching, and is fully containerised with Docker Compose.

---

## Table of Contents

- [Stack](#stack)
- [Project Structure](#project-structure)
- [Setup](#setup)
  - [Prerequisites](#prerequisites)
  - [Running with Docker](#running-with-docker)
  - [Running locally](#running-locally)
  - [Running tests](#running-tests)
- [API Reference](#api-reference)
- [Interactive Docs](#interactive-docs)
- [Design Choices](#design-choices)
- [What I'd Improve with More Time](#what-id-improve-with-more-time)

---

## Stack

| Concern        | Choice                          |
| -------------- | ------------------------------- |
| Language       | Go 1.23+                        |
| HTTP framework | Gin                             |
| Database       | PostgreSQL 16 + PostGIS 3.4     |
| Cache          | Redis 7                         |
| Migrations     | golang-migrate                  |
| Validation     | go-playground/validator v10     |
| Logging        | logrus (structured JSON)        |
| API docs       | Swagger UI (swaggo/gin-swagger) |
| Testing        | testify + pgx pool              |
| Containers     | Docker + Docker Compose         |
| Reverse proxy  | nginx                           |
| CI/CD          | GitHub Actions                  |

---

## Project Structure

```
.
├── cmd/api/              # Application entry point
├── configs/              # application.yaml — single source of config
├── deployments/          # Dockerfile, docker-compose.yml, nginx.conf
├── docs/                 # Auto-generated Swagger docs (swag init)
├── internal/
│   ├── agent/            # Agent domain, repository, service
│   ├── health/           # Health check handler
│   ├── listing/
│   │   ├── application/  # One use-case per operation (create, get, update, delete, search)
│   │   ├── domain/       # Listing entity, ListingType enum, SearchFilters
│   │   ├── handler/      # HTTP handlers, request/response types
│   │   └── repository/   # Repository interface + PostgreSQL implementation
│   └── search/           # Search service with Redis cache-aside
├── migrations/           # SQL migration files (up + down)
├── pkg/
│   ├── cache/            # Redis client wrapper
│   ├── database/         # pgxpool setup + migration runner
│   ├── httpx/            # Standardised error and response helpers
│   ├── logger/           # logrus wrapper
│   ├── pagination/       # Page/per_page parsing and Meta struct
│   └── validation/       # Struct validation with human-readable messages
└── tests/
    ├── fixtures/         # SQL seed data
    └── integration/      # Integration tests against a real database
```

---

## Setup

### Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (includes Docker Compose)
- Go 1.23+ — only needed for local development outside Docker

### Running with Docker

All configuration lives in `configs/application.yaml`. Open it and set your passwords before starting:

```yaml
database:
  password: changeme # must match POSTGRES_PASSWORD in docker-compose.yml

redis:
  password: redispass # must match the --requirepass value in docker-compose.yml
```

Then start everything:

```bash
docker compose -f deployments/docker-compose.yml up -d --build
```

This starts four containers — `postgres` (PostGIS), `redis`, `api`, and `nginx`. Database migrations run automatically on startup.

| Service      | URL                                 |
| ------------ | ----------------------------------- |
| API (nginx)  | http://localhost                    |
| API (direct) | http://localhost:8080               |
| Swagger UI   | http://localhost/swagger/index.html |
| Health       | http://localhost/health             |

To stop and remove all containers including data volumes:

```bash
docker compose -f deployments/docker-compose.yml down -v
```

> **Note:** If you change the database password after the volume has been created, run `down -v` first to wipe the volume and let postgres reinitialise with the new password.

### Running locally

Start only the infrastructure:

```bash
docker compose -f deployments/docker-compose.yml up -d postgres redis
```

Update `configs/application.yaml` to point at localhost:

```yaml
database:
  host: localhost

redis:
  host: localhost
```

Then run the API:

```bash
make run
```

### Running tests

Integration tests require a running PostgreSQL instance with PostGIS. The easiest way is to use the docker-compose postgres service.

```bash
# Unit tests
make test

# Integration tests
DB_HOST=localhost DB_NAME=property_marketplace_test DB_PASSWORD=changeme make test-integration
```

To regenerate Swagger docs after changing handler annotations:

```bash
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
```

---

## API Reference

Base path: `/api/v1`

All successful responses wrap data in a `data` field. Paginated responses also include a `pagination` object.

### Health

```
GET /health
```

```json
{ "status": "ok", "database": "ok", "timestamp": "2026-09-29T11:00:00Z" }
```

### Listings

| Method   | Path               | Description                   |
| -------- | ------------------ | ----------------------------- |
| `POST`   | `/listings`        | Create a listing              |
| `GET`    | `/listings`        | List all listings (paginated) |
| `GET`    | `/listings/:id`    | Get a single listing by UUID  |
| `PUT`    | `/listings/:id`    | Update a listing              |
| `DELETE` | `/listings/:id`    | Delete a listing              |
| `GET`    | `/listings/search` | Filter and geo-radius search  |

#### Create / Update body

```json
{
  "title": "3-Bed Flat in Lekki",
  "description": "Spacious and modern",
  "price": 1500000,
  "type": "rent",
  "bedrooms": 3,
  "address": "14 Admiralty Way, Lekki Phase 1",
  "latitude": 6.4281,
  "longitude": 3.4219,
  "agent_id": "a1000000-0000-0000-0000-000000000001"
}
```

`type` must be one of: `rent`, `sale`, `shortlet`.

#### Search query parameters

| Param       | Type   | Description                                       |
| ----------- | ------ | ------------------------------------------------- |
| `type`      | string | Filter by listing type (`rent`/`sale`/`shortlet`) |
| `min_price` | float  | Minimum price (inclusive)                         |
| `max_price` | float  | Maximum price (inclusive)                         |
| `bedrooms`  | int    | Exact bedroom count                               |
| `lat`       | float  | Latitude of search origin                         |
| `lng`       | float  | Longitude of search origin                        |
| `radius_km` | float  | Radius in kilometres — requires `lat` and `lng`   |
| `page`      | int    | Page number (default: 1)                          |
| `per_page`  | int    | Results per page (default: 20, max: 100)          |

`lat`, `lng` and `radius_km` must all be provided together for geo search.

#### Paginated response shape

```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total_items": 42,
    "total_pages": 3
  }
}
```

#### Error response shape

```json
{ "message": "listing not found" }
```

Validation errors return HTTP 422:

```json
{
  "errors": [
    { "field": "price", "message": "must be greater than 0" },
    { "field": "type", "message": "must be one of: rent sale shortlet" }
  ]
}
```

---

## Interactive Docs

Swagger UI is served at:

```
http://localhost/swagger/index.html
```

All endpoints are documented with example request bodies and response schemas. You can create agents directly via the postgres container to get a valid `agent_id` for testing:

```bash
docker exec -it property_postgres psql -U postgres -d property_marketplace \
  -c "INSERT INTO agents (id, name, email, phone, agency) \
      VALUES (gen_random_uuid(), 'Test Agent', 'agent@test.com', '+2348000000000', 'Test Realty') \
      RETURNING id;"
```

---

## Design Choices

**Modular monolith, not microservices.**
All modules live in one deployable binary but are separated by Go packages (`internal/listing`, `internal/agent`, `internal/search`). This keeps operational complexity low while enforcing clear boundaries that could be extracted into services later without a significant rewrite.

**Use-case layer per operation.**
Each action (create, get, update, delete, search) is its own struct with a single `Execute` method. This keeps HTTP handlers thin — they translate HTTP concerns into inputs and delegate all logic to the use-case. It also makes business rules unit-testable without needing a real HTTP server.

**Repository interfaces.**
Every handler and use-case depends on an interface, never a concrete struct. This makes it straightforward to swap the PostgreSQL implementation for an in-memory fake in tests, and documents the data contract explicitly.

**PostGIS `GEOGRAPHY` column for geospatial queries.**
The `location` column is a `GENERATED ALWAYS AS ST_MakePoint(longitude, latitude)::geography STORED` computed column, derived automatically from the stored lat/lng values. Queries use `ST_DWithin` which works in metres on the WGS-84 spheroid — this is more accurate than planar distance calculations for real-world distances. A GIST index on the column keeps radius searches fast even at scale.

**Redis cache-aside for search results.**
Search results are cached for 5 minutes keyed by an MD5 hash of the serialised filter and pagination parameters. The cache client is optional — if Redis is unavailable at startup the API logs a warning and continues without caching. This prevents a Redis outage from taking down the whole service.

**YAML-only configuration.**
All configuration (database, Redis, server) is read from `configs/application.yaml`. There are no required environment variables. This makes the service easy to reason about locally and in containers — change one file, restart the container.

**Graceful shutdown.**
On `SIGINT` or `SIGTERM` the HTTP server stops accepting new connections and waits up to 10 seconds for in-flight requests to complete before exiting. This prevents dropped requests during rolling deploys.

**Non-root Docker image.**
The production stage runs as a dedicated `appuser` with no shell, keeping the attack surface minimal. The build uses a multi-stage Dockerfile so the final image contains only the compiled binary, config, and migrations — no Go toolchain.

---

## What I'd Improve with More Time

**Agent HTTP endpoints.**
The agent domain, repository, and service are fully implemented. The missing piece is an HTTP handler with routes for `POST /agents`, `GET /agents/:id`, `PUT /agents/:id`, `DELETE /agents/:id`. It follows exactly the same pattern as the listing handler and would take under an hour to add.

**Authentication and authorisation.**
`golang-jwt/jwt` is already in `go.mod`. A JWT middleware would authenticate requests and scope write operations (create/update/delete listing) to the agent who owns them. Without this, any caller can mutate any listing.

**Cursor-based pagination.**
The current `LIMIT / OFFSET` approach degrades in performance as offsets grow large because the database still scans and discards the skipped rows. A keyset cursor using `(created_at, id)` would keep pagination O(1) regardless of page depth.

**Full-text search on title and description.**
A `tsvector` generated column with a GIN index would enable efficient keyword search within PostgreSQL. For more advanced relevance ranking, faceted search, and typo tolerance, integrating Meilisearch would be the next step — the architecture already isolates search behind its own service, so swapping the backend would not touch the handlers.

**Observability.**
The dependency tree already includes OpenTelemetry and logrus. Wiring up distributed tracing (OTLP exporter), structured request logs with trace IDs, and a Prometheus `/metrics` endpoint would give full production visibility.

**Rate limiting.**
The search endpoint with a geo query and no rate limit is a potential abuse vector. A token-bucket middleware per IP (or per agent JWT) would protect it. `golang.org/x/time/rate` is in the stdlib ecosystem and requires no new dependency.

**Contract testing.**
The Swagger spec is already generated. Running `schemathesis` or `oapi-codegen` against the spec in CI would catch any drift between the documented and actual API behaviour automatically.

**Structured config validation.**
Currently a missing or zero-valued config field silently produces a broken connection string. Adding validation at startup (required fields, port ranges, non-empty passwords) would surface misconfiguration immediately rather than at the first database call.
