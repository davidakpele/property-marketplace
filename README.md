# Property Marketplace API

A property listing REST API built with Go, PostgreSQL + PostGIS, and Redis. It supports full CRUD for listings and agents, geospatial radius search, Redis-backed caching, and is fully containerised with Docker Compose behind a production-grade nginx reverse proxy.

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
- [Nginx Architecture](#nginx-architecture)
- [Query Logging](#query-logging)
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
| Reverse proxy  | nginx 1.27 (multi-file config)  |
| CI/CD          | GitHub Actions                  |

---

## Project Structure

```
.
├── cmd/api/              # Application entry point
├── configs/              # application.yaml — single source of config
├── deployments/
│   ├── Dockerfile        # Multi-stage Go build → alpine runtime
│   ├── docker-compose.yml
│   └── nginx/
│       ├── Dockerfile    # nginx image built from conf.d directory
│       ├── nginx.conf    # Main nginx config (worker tuning, gzip, log formats)
│       └── conf.d/
│           ├── 01-security-headers.conf   # CSP, X-Frame-Options, HSTS headers
│           ├── 02-rate-limiting.conf      # Per-zone rate limit definitions
│           ├── 03-bot-detection.conf      # Malicious request and UA maps
│           ├── 04-upstreams.conf          # property_api upstream (keepalive)
│           ├── 05-server-http.conf        # Main server block with all locations
│           ├── 06-server-https.conf       # TLS block (commented, ready to enable)
│           ├── 07-error-pages.conf        # JSON error responses for all 4xx/5xx
│           ├── 08-security-blocks.conf    # Attack pattern location blocks
│           ├── 10-health-checks.conf      # /nginx-health and /backend-health
│           ├── 11-content-cache.conf      # Cache bypass maps
│           ├── globalblacklist.conf       # Bad bot user-agent blocklist
│           └── swagger-locations.conf     # /swagger/ proxy + redirect rules
├── docs/                 # Auto-generated Swagger docs (swag init)
├── internal/
│   ├── agent/            # Agent domain, repository, handler, service
│   ├── health/           # Health check handler
│   ├── listing/
│   │   ├── application/  # One use-case per operation
│   │   ├── domain/       # Listing entity, ListingType enum, SearchFilters, errors
│   │   ├── handler/      # HTTP handlers, request/response types
│   │   └── repository/   # Repository interface + PostgreSQL implementation
│   └── search/           # Search service with Redis cache-aside
├── migrations/           # SQL migration files (up + down)
├── pkg/
│   ├── cache/            # Redis client wrapper
│   ├── database/         # pgxpool setup, migration runner, query tracer
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

All configuration lives in `configs/application.yaml`. Open it and confirm the passwords match the values in `docker-compose.yml`:

```yaml
database:
  password: changeme # must match POSTGRES_PASSWORD in docker-compose.yml

redis:
  password: redispass # must match --requirepass in docker-compose.yml
```

Start everything:

```bash
docker compose -f deployments/docker-compose.yml up -d --build
```

This starts four containers — `postgres` (PostGIS), `redis`, `api`, and `nginx`. Database migrations run automatically on startup.

| Service        | URL                                 |
| -------------- | ----------------------------------- |
| API via nginx  | http://localhost                    |
| API direct     | http://localhost:8080               |
| Swagger UI     | http://localhost/swagger/index.html |
| Health (nginx) | http://localhost/health             |
| Nginx health   | http://localhost/nginx-health       |
| Backend health | http://localhost/backend-health     |

To stop and remove all containers including data volumes:

```bash
docker compose -f deployments/docker-compose.yml down -v
```

> **Note:** If you change the database password after the volume was first created, run `down -v` to wipe the volume so postgres reinitialises with the new password.

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

Run the API:

```bash
make run
```

### Running tests

Integration tests require a running PostgreSQL instance with PostGIS.

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

All successful responses wrap data in a `data` field. Paginated responses include a `pagination` object.

### Health

```
GET /health
```

```json
{ "status": "ok", "database": "ok", "timestamp": "2026-09-29T11:00:00Z" }
```

### Agents

| Method   | Path          | Description                 |
| -------- | ------------- | --------------------------- |
| `POST`   | `/agents`     | Create an agent             |
| `GET`    | `/agents`     | List all agents (paginated) |
| `GET`    | `/agents/:id` | Get a single agent by UUID  |
| `PUT`    | `/agents/:id` | Update an agent             |
| `DELETE` | `/agents/:id` | Delete an agent             |

#### Create / Update agent body

```json
{
  "name": "Ada Okafor",
  "email": "ada@realty.ng",
  "phone": "+2348011111111",
  "agency": "Realty NG"
}
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

Swagger UI is available at:

```
http://localhost/swagger/index.html
```

All endpoints are documented with example request bodies and response schemas. Create an agent first to get a valid `agent_id`, then use that UUID when creating listings.

---

## Nginx Architecture

The nginx setup is built as its own Docker image (`deployments/nginx/Dockerfile`) with a modular configuration split across numbered files in `conf.d/`. Each file has a single responsibility, making it easy to adjust one concern without touching the others.

### Configuration files

| File                       | Purpose                                                                                                                                                                                                                                                                             |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `nginx.conf`               | Worker tuning, connection limits, gzip, JSON log formats, global proxy settings, circuit breaker (`proxy_next_upstream`)                                                                                                                                                            |
| `01-security-headers.conf` | Adds `X-Frame-Options`, `X-Content-Type-Options`, `X-XSS-Protection`, `Referrer-Policy`, `Content-Security-Policy`, `Permissions-Policy`, and a custom `X-Property-API-Version` header to every response. Strips `X-Powered-By`, `X-Runtime`, and `Server` from upstream responses. |
| `02-rate-limiting.conf`    | Defines per-IP rate limit zones: `api_limit` (30 req/s), `search_limit` (10 req/s — tighter because geo queries are more expensive), `strict_limit` (5 req/s), `global_limit` (100 req/s). Also defines connection limit zones.                                                     |
| `03-bot-detection.conf`    | Three nginx `map` directives: `$is_malicious` (detects SQL injection, XSS, path traversal, null bytes in the request URI), `$is_suspicious` (matches scanner and scraper user-agent strings), `$admin_access` (geo-based allowlist for internal networks).                          |
| `04-upstreams.conf`        | Defines the `property_api` upstream pointing at `api:8080` with keepalive connection pooling (32 connections, 1000 requests, 60 s timeout).                                                                                                                                         |
| `05-server-http.conf`      | Main server block on port 80. Includes error pages and security blocks, then defines location blocks in priority order: `/health`, health checks, Swagger, `/api/v1/listings/search` (search-specific rate limit), `/api/v1/` (general API), catch-all JSON 404.                    |
| `06-server-https.conf`     | TLS server block, fully commented out. Ready to enable by uncommenting and providing certificate paths.                                                                                                                                                                             |
| `07-error-pages.conf`      | Named locations (`@bad_request`, `@unauthorized`, `@forbidden`, `@not_found`, `@too_many_requests`, `@service_unavailable`, etc.) that return structured JSON for every 4xx and 5xx status code, including a `request_id` field for correlation.                                    |
| `08-security-blocks.conf`  | Location blocks that return 403 or 404 for hidden files (`.env`, `.git`), backup file extensions (`.bak`, `.log`), admin panel probing (`phpmyadmin`, `wp-login`), SQL injection patterns, XSS patterns, and path traversal patterns.                                               |
| `10-health-checks.conf`    | `/nginx-health` returns 200 immediately (no upstream needed). `/backend-health` proxies to the API's `/health` endpoint, restricted to internal IP ranges only.                                                                                                                     |
| `11-content-cache.conf`    | Maps that control cache bypass: non-GET methods always bypass, `Cache-Control: no-cache/no-store` headers bypass.                                                                                                                                                                   |
| `globalblacklist.conf`     | Comprehensive bad-bot user-agent blocklist (600+ entries) sourced from community-maintained lists. Maps agents to block scores — score 3 is blocked, score 0 is a known good bot (Googlebot, Bingbot, etc.).                                                                        |
| `swagger-locations.conf`   | `/swagger` and `/docs` redirect to `/swagger/index.html`. `/swagger/*` is proxied to the API with a 1-hour cache header. Malicious request check applied before proxying.                                                                                                           |

### Rate limiting zones

| Zone           | Rate      | Used on                                           |
| -------------- | --------- | ------------------------------------------------- |
| `api_limit`    | 30 req/s  | All `/api/v1/` requests (burst 50)                |
| `search_limit` | 10 req/s  | `/api/v1/listings/search` specifically (burst 20) |
| `strict_limit` | 5 req/s   | Available for future sensitive endpoints          |
| `global_limit` | 100 req/s | Global safety net                                 |

### Security layers

Requests pass through multiple independent layers before reaching the API:

1. **Bot blocklist** — user-agent matched against `globalblacklist.conf` before any location block is evaluated
2. **Attack pattern detection** — `$is_malicious` map checked inside each API location; returns 403 with a structured error before proxying
3. **Security location blocks** — file-extension and URL-pattern blocks in `08-security-blocks.conf` intercept known attack vectors
4. **Rate limiting** — per-IP limits applied at the location level, returning 429 with `Retry-After` header via `07-error-pages.conf`
5. **Security headers** — applied globally via `01-security-headers.conf`, upstream response headers stripped

### Health endpoints

| Endpoint              | Access            | Description                                                   |
| --------------------- | ----------------- | ------------------------------------------------------------- |
| `GET /health`         | Public            | Proxied to the API health check                               |
| `GET /nginx-health`   | Public            | nginx self-check, no upstream needed                          |
| `GET /backend-health` | Internal IPs only | Proxied to `api:8080/health`, returns 503 JSON if API is down |

### CORS

CORS preflight (`OPTIONS`) requests are handled inline within the API location blocks — no additional middleware needed. Allowed methods: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`. The `Origin` header is echoed back.

### TLS

`06-server-https.conf` contains a fully commented-out TLS server block ready to activate. To enable HTTPS:

1. Place your certificate at `/etc/nginx/ssl/fullchain.pem` and key at `/etc/nginx/ssl/privkey.pem` inside the nginx container
2. Uncomment the server block in `06-server-https.conf`
3. Add a redirect from port 80 to 443 in `05-server-http.conf`
4. Rebuild the nginx container

---

## Query Logging

SQL queries can be logged directly to the API container logs for debugging. Toggle it in `configs/application.yaml`:

```yaml
database:
  log_queries: true # set to false in production
```

When enabled, every query is logged as a structured JSON line:

```json
{
  "level": "debug",
  "msg": "query executed",
  "sql": "SELECT id, title ... FROM listings WHERE ...",
  "args": [6.4281, 3.4219, 5000],
  "duration": "1.8ms",
  "rows": 3
}
```

Failed queries log at `error` level with an `error` field. The tracer uses pgx's native `Tracer` interface — zero overhead when disabled, no third-party dependency needed.

Tail the logs:

```bash
docker logs property_api -f
```

---

## Design Choices

**Modular monolith, not microservices.**
All modules live in one deployable binary but are separated by Go packages (`internal/listing`, `internal/agent`, `internal/search`). This keeps operational complexity low while enforcing clear boundaries that could be extracted into services later without a significant rewrite.

**Use-case layer per operation.**
Each action (create, get, list, update, delete, search) is its own struct with a single `Execute` method. Handlers translate HTTP concerns into inputs and delegate all logic to the use-case. This keeps handlers thin and makes business rules unit-testable without a real HTTP server.

**Repository interfaces.**
Every handler and use-case depends on an interface, never a concrete struct. This makes it straightforward to swap the PostgreSQL implementation for an in-memory fake in tests, and documents the data contract explicitly.

**Domain sentinel errors.**
Repositories return domain errors (`domain.ErrListingNotFound`, `domain.ErrAgentNotFound`) rather than leaking database-layer errors (`pgx.ErrNoRows`) into the business logic. Use-cases check domain errors only — no pgx imports above the repository layer.

**PostGIS `GEOGRAPHY` column for geospatial queries.**
The `location` column is `GENERATED ALWAYS AS ST_MakePoint(longitude, latitude)::geography STORED` — computed automatically from the stored lat/lng. Queries use `ST_DWithin` which works in metres on the WGS-84 spheroid, more accurate than planar distance for real-world radii. A GIST index keeps radius searches fast at scale.

**Redis cache-aside for search results.**
Search results are cached for 5 minutes keyed by an MD5 hash of the serialised filter and pagination parameters. If Redis is unavailable at startup the API logs a warning and continues without caching — a Redis outage cannot take down the API.

**YAML-only configuration.**
All configuration lives in `configs/application.yaml`. No required environment variables. Change one file, restart the container.

**Graceful shutdown.**
On `SIGINT` or `SIGTERM` the HTTP server stops accepting connections and waits up to 10 seconds for in-flight requests before exiting. No dropped requests during deploys.

**Non-root Docker image.**
The production stage runs as a dedicated `appuser` with no shell. Multi-stage build means the final image contains only the binary, config, and migrations — no Go toolchain.

**Modular nginx over a single flat config.**
Each nginx concern is isolated in its own numbered file. Security headers, rate limiting, bot detection, upstreams, error pages, and attack blocks can each be adjusted independently without risk of breaking an unrelated rule.

---

## What I'd Improve with More Time

**Authentication and authorisation.**
`golang-jwt/jwt` is already in `go.mod`. A JWT middleware would authenticate requests and scope write operations (create/update/delete) to the owning agent.

**Cursor-based pagination.**
`LIMIT / OFFSET` degrades at large offsets — the database scans and discards skipped rows. A keyset cursor on `(created_at, id)` keeps pagination O(1) at any depth.

**Full-text search.**
A `tsvector` generated column with a GIN index enables keyword search within PostgreSQL. For relevance ranking, typo tolerance, and faceted filters, Meilisearch is the next step — the architecture already isolates search behind its own service, so swapping the backend would not touch the handlers.

**Observability.**
OpenTelemetry and logrus are already in the dependency tree. Wiring up distributed tracing (OTLP exporter), structured request logs with trace IDs, and a Prometheus `/metrics` endpoint would give full production visibility.

**Per-user rate limiting.**
Nginx rate limits by IP. Once JWT auth is in place, rate limiting per authenticated user (or per agent tier) would be more accurate and harder to bypass.

**Contract testing.**
The Swagger spec is already generated. Running `schemathesis` against it in CI would catch drift between the documented and actual API behaviour automatically.

**Structured config validation.**
A missing or zero-valued config field silently produces a broken DSN. Validating required fields at startup would surface misconfiguration immediately.
