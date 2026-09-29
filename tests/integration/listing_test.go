package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	listing "github.com/davidakpele/property-marketplace/internal/listing"
	listinghandler "github.com/davidakpele/property-marketplace/internal/listing/handler"
	listingrepo "github.com/davidakpele/property-marketplace/internal/listing/repository"
	"github.com/davidakpele/property-marketplace/internal/search"
	"github.com/davidakpele/property-marketplace/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	cfg := database.Config{
		Host:     envOr("DB_HOST", "localhost"),
		Port:     5432,
		Name:     envOr("DB_NAME", "property_marketplace_test"),
		User:     envOr("DB_USER", "postgres"),
		Password: envOr("DB_PASSWORD", "testpassword"),
		SSLMode:  envOr("DB_SSL_MODE", "disable"),
	}
	pool, err := database.NewPool(context.Background(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })

	require.NoError(t, database.RunMigrations(cfg.DSN(), "../../migrations"))
	return pool
}

func testRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	repo := listingrepo.NewPostgresListingRepository(pool)
	svc := listing.NewService(repo)
	searchSvc := search.NewService(repo, nil)

	api := r.Group("/api/v1")
	listinghandler.NewHandler(svc, searchSvc).RegisterRoutes(api)
	return r
}

func cleanListings(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), "DELETE FROM listings")
	require.NoError(t, err)
	_, err = pool.Exec(context.Background(), "DELETE FROM agents")
	require.NoError(t, err)
}

func seedAgent(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO agents (id, name, email, phone, agency, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`,
		id, "Test Agent", fmt.Sprintf("agent-%s@test.com", id), "+2348099999999", "Test Agency",
	)
	require.NoError(t, err)
	return id
}

func TestCreateListing(t *testing.T) {
	pool := testPool(t)
	cleanListings(t, pool)
	agentID := seedAgent(t, pool)
	r := testRouter(pool)

	body := map[string]interface{}{
		"title":     "Nice 2-Bed Flat",
		"price":     750000,
		"type":      "rent",
		"bedrooms":  2,
		"address":   "10 Test Street, Lagos",
		"latitude":  6.4281,
		"longitude": 3.4219,
		"agent_id":  agentID.String(),
	}

	w := doRequest(t, r, http.MethodPost, "/api/v1/listings", body)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Nice 2-Bed Flat", data["title"])
	assert.Equal(t, "rent", data["type"])
	assert.NotEmpty(t, data["id"])
}

func TestCreateListing_ValidationError(t *testing.T) {
	pool := testPool(t)
	r := testRouter(pool)

	body := map[string]interface{}{
		"price":    -100,
		"type":     "unknown",
		"bedrooms": 2,
	}

	w := doRequest(t, r, http.MethodPost, "/api/v1/listings", body)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestGetListing(t *testing.T) {
	pool := testPool(t)
	cleanListings(t, pool)
	agentID := seedAgent(t, pool)
	r := testRouter(pool)

	created := createTestListing(t, r, agentID)
	id := created["id"].(string)

	w := doRequest(t, r, http.MethodGet, "/api/v1/listings/"+id, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, id, data["id"])
}

func TestGetListing_NotFound(t *testing.T) {
	pool := testPool(t)
	r := testRouter(pool)

	w := doRequest(t, r, http.MethodGet, "/api/v1/listings/"+uuid.New().String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateListing(t *testing.T) {
	pool := testPool(t)
	cleanListings(t, pool)
	agentID := seedAgent(t, pool)
	r := testRouter(pool)

	created := createTestListing(t, r, agentID)
	id := created["id"].(string)

	update := map[string]interface{}{
		"title":     "Updated Title",
		"price":     900000,
		"type":      "sale",
		"bedrooms":  3,
		"address":   "20 Updated St, Lagos",
		"latitude":  6.4281,
		"longitude": 3.4219,
		"agent_id":  agentID.String(),
	}

	w := doRequest(t, r, http.MethodPut, "/api/v1/listings/"+id, update)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Updated Title", data["title"])
	assert.Equal(t, "sale", data["type"])
}

func TestDeleteListing(t *testing.T) {
	pool := testPool(t)
	cleanListings(t, pool)
	agentID := seedAgent(t, pool)
	r := testRouter(pool)

	created := createTestListing(t, r, agentID)
	id := created["id"].(string)

	w := doRequest(t, r, http.MethodDelete, "/api/v1/listings/"+id, nil)
	assert.Equal(t, http.StatusNoContent, w.Code)

	w2 := doRequest(t, r, http.MethodGet, "/api/v1/listings/"+id, nil)
	assert.Equal(t, http.StatusNotFound, w2.Code)
}

func TestListListings_Pagination(t *testing.T) {
	pool := testPool(t)
	cleanListings(t, pool)
	agentID := seedAgent(t, pool)
	r := testRouter(pool)

	for i := 0; i < 5; i++ {
		createTestListing(t, r, agentID)
	}

	w := doRequest(t, r, http.MethodGet, "/api/v1/listings?page=1&per_page=3", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	data := resp["data"].([]interface{})
	assert.Len(t, data, 3)

	pagination := resp["pagination"].(map[string]interface{})
	assert.Equal(t, float64(5), pagination["total_items"])
	assert.Equal(t, float64(2), pagination["total_pages"])
}

func createTestListing(t *testing.T, r *gin.Engine, agentID uuid.UUID) map[string]interface{} {
	t.Helper()
	body := map[string]interface{}{
		"title":     "Test Listing",
		"price":     500000,
		"type":      "rent",
		"bedrooms":  2,
		"address":   "1 Test Rd, Lagos",
		"latitude":  6.4281,
		"longitude": 3.4219,
		"agent_id":  agentID.String(),
	}
	w := doRequest(t, r, http.MethodPost, "/api/v1/listings", body)
	require.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp["data"].(map[string]interface{})
}

func doRequest(t *testing.T, r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var b bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&b).Encode(body))
	}
	req, err := http.NewRequest(method, path, &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
