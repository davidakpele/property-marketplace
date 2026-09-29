# Property Marketplace API

A production-ready property listing REST API built with Go, PostgreSQL + PostGIS, and Redis.

## Stack

| Concern        | Choice                              |
| -------------- | ----------------------------------- |
| Language       | Go 1.23                             |
| HTTP framework | Gin                                 |
| Database       | PostgreSQL 16 + PostGIS 3.4         |
| Cache          | Redis 7                             |
| Migrations     | golang-migrate                      |
| Validation     | go-playground/validator             |
| Logging        | logrus (JSON)                       |
| Testing        | testify + testcontainers (pgx pool) |
| Container      | Docker + docker compose             |
| Reverse proxy  | nginx                               |
| CI/CD          | GitHub Actions                      |

## Setup

### Prerequisites

- Docker and Docker Compose
- Go 1.23+ (for local development)

### Quick start with Docker

```bash
cp .env.example .env
# edit .env with real passwords

make docker-up
```

The API will be available at `http://localhost` (nginx on port 80) or directly at `http://localhost:8080`.

### Local development

1. Start infrastructure only:

```bash
docker compose -f deployments/docker-compose.yml up -d postgres redis
```

2. Copy and configure env:

```bash
cp .env.example .env
```

3. Run the API:

```bash
make run
```

### Running tests

Integration tests require a running PostgreSQL with PostGIS.

```bash
# unit tests only
make test

# integration tests (needs DB)
DB_HOST=localhost DB_PASSWORD=testpassword make test-integration
```

## API Reference

Base path: `/api/v1`

### Health

```
GET /health
```

### Listings

| Method | Path               | Description                      |
| ------ | ------------------ | -------------------------------- |
| POST   | `/listings`        | Create a listing                 |
| GET    | `/listings`        | List all listings (paginated)    |
| GET    | `/listings/:id`    | Get a single listing             |
| PUT    | `/listings/:id`    | Update a listing                 |
| DELETE | `/listings/:id`    | Delete a listing                 |
| GET    | `/listings/search` | Search with filters + geo radius |

#### Create / Update listing body

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

| Param       | Type   | Description                              |
| ----------- | ------ | ---------------------------------------- |
| `type`      | string | Filter by listing type                   |
| `min_price` | float  | Minimum price                            |
| `max_price` | float  | Maximum price                            |
| `bedrooms`  | int    | Exact bedroom count                      |
| `lat`       | float  | Latitude of search origin                |
| `lng`       | float  | Longitude of search origin               |
| `radius_km` | float  | Search radius in kilometres              |
| `page`      | int    | Page number (default: 1)                 |
| `per_page`  | int    | Results per page (default: 20, max: 100) |

`lat`, `lng` and `radius_km` must all be supplied together for geo search.

#### Paginated response shape

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total_items": 42,
    "total_pages": 3
  }
}
```

## Design Choices

**Modular monolith over microservices.** At this scale a monolith is simpler to operate, test, and reason about. Module boundaries (`internal/listing`, `internal/agent`, `internal/search`) are enforced by Go packages so the code can be split into services later without a rewrite.

**Use-case layer per operation.** Each action (create, get, update, delete, search) is its own struct with an `Execute` method. This keeps handlers thin, makes business logic independently testable, and mirrors the Command pattern.

**Repository interfaces.** Handlers and use-cases depend only on interfaces, making it straightforward to swap implementations or write in-memory fakes for unit tests.

**PostGIS `GEOGRAPHY` column for geo search.** The `location` column is a `GENERATED ALWAYS AS` computed geography from `latitude`/`longitude`. Queries use `ST_DWithin` which operates in metres on the spheroid — accurate for real-world radius searches. A GIST index keeps it fast.

**Redis cache-aside for search.** Search results are cached for 5 minutes using an MD5 of the serialised filter + pagination params as the key. The cache is optional — the app degrades gracefully if Redis is unavailable.

**Graceful shutdown.** The HTTP server waits up to 10 seconds for in-flight requests before exiting, preventing connection drops on deploy.

**Non-root Docker image.** The production image runs as a dedicated `appuser`, reducing the blast radius of any container escape.

## What I'd Improve with More Time

- **Agent HTTP handler** — the agent service is fully wired but has no HTTP routes. Adding `POST /agents`, `GET /agents/:id`, etc. follows the same pattern as listings.
- **Authentication** — JWT middleware (`golang-jwt/jwt` is already in `go.mod`) to scope listing mutations to the owning agent.
- **Cursor-based pagination** — offset pagination degrades at large offsets; a `created_at + id` cursor would be more efficient.
- **Full-text search** — add a `tsvector` column and GIN index for keyword search on title/description, or integrate Meilisearch for relevance ranking.
- **Structured config validation** — use `caarlos0/env` (already in `go.mod`) to validate required env vars at startup rather than silent zero-values.
- **Observability** — wire up OpenTelemetry traces and Prometheus metrics (both deps are present in `go.mod`).
- **Rate limiting** — add a token-bucket middleware per IP to protect the search endpoint.
- **Contract tests** — generate an OpenAPI spec and validate handler inputs/outputs against it in CI.
