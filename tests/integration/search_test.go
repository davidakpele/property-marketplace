package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertListing(t *testing.T, pool *pgxpool.Pool, agentID uuid.UUID, title, ltype string, price float64, bedrooms int, lat, lng float64) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO listings (id, title, description, price, type, bedrooms, address, latitude, longitude, agent_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())`,
		uuid.New(), title, "", price, ltype, bedrooms,
		fmt.Sprintf("%s address", title), lat, lng, agentID,
	)
	require.NoError(t, err)
}

func TestSearch_ByType(t *testing.T) {
	pool := testPool(t)
	cleanListings(t, pool)
	agentID := seedAgent(t, pool)
	r := testRouter(pool)

	insertListing(t, pool, agentID, "rent listing",  "rent",     500000,  2, 6.4281, 3.4219)
	insertListing(t, pool, agentID, "sale listing",  "sale",     3000000, 3, 6.4280, 3.4296)
	insertListing(t, pool, agentID, "shortlet flat", "shortlet", 80000,   1, 6.5095, 3.3711)

	w := doRequest(t, r, http.MethodGet, "/api/v1/listings/search?type=rent", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
	assert.Equal(t, "rent listing", data[0].(map[string]interface{})["title"])
}

func TestSearch_ByPriceRange(t *testing.T) {
	pool := testPool(t)
	cleanListings(t, pool)
	agentID := seedAgent(t, pool)
	r := testRouter(pool)

	insertListing(t, pool, agentID, "cheap",     "rent", 100000,  1, 6.4281, 3.4219)
	insertListing(t, pool, agentID, "mid",       "rent", 500000,  2, 6.4280, 3.4296)
	insertListing(t, pool, agentID, "expensive", "rent", 5000000, 4, 6.5095, 3.3711)

	w := doRequest(t, r, http.MethodGet, "/api/v1/listings/search?min_price=200000&max_price=1000000", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
	assert.Equal(t, "mid", data[0].(map[string]interface{})["title"])
}

func TestSearch_ByBedrooms(t *testing.T) {
	pool := testPool(t)
	cleanListings(t, pool)
	agentID := seedAgent(t, pool)
	r := testRouter(pool)

	insertListing(t, pool, agentID, "1-bed", "rent", 200000, 1, 6.4281, 3.4219)
	insertListing(t, pool, agentID, "2-bed", "rent", 300000, 2, 6.4280, 3.4296)
	insertListing(t, pool, agentID, "3-bed", "rent", 400000, 3, 6.5095, 3.3711)

	w := doRequest(t, r, http.MethodGet, "/api/v1/listings/search?bedrooms=2", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
	assert.Equal(t, "2-bed", data[0].(map[string]interface{})["title"])
}

func TestSearch_GeoRadius(t *testing.T) {
	pool := testPool(t)
	cleanListings(t, pool)
	agentID := seedAgent(t, pool)
	r := testRouter(pool)

	insertListing(t, pool, agentID, "nearby in Lekki",   "rent", 500000, 2, 6.4281, 3.4219)
	insertListing(t, pool, agentID, "near VI",           "rent", 600000, 2, 6.4280, 3.4296)
	insertListing(t, pool, agentID, "far away in Abuja", "rent", 700000, 2, 9.0820, 7.4891)

	url := "/api/v1/listings/search?lat=6.4281&lng=3.4219&radius_km=5"
	w := doRequest(t, r, http.MethodGet, url, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	assert.Len(t, data, 2, "should only return Lagos listings within 5 km, not Abuja")
}

func TestSearch_GeoRadius_MissingParams(t *testing.T) {
	pool := testPool(t)
	r := testRouter(pool)

	w := doRequest(t, r, http.MethodGet, "/api/v1/listings/search?lat=6.4281&lng=3.4219", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearch_CombinedFilters(t *testing.T) {
	pool := testPool(t)
	cleanListings(t, pool)
	agentID := seedAgent(t, pool)
	r := testRouter(pool)

	insertListing(t, pool, agentID, "match",    "sale", 2000000, 3, 6.4281, 3.4219)
	insertListing(t, pool, agentID, "no-type",  "rent", 2000000, 3, 6.4280, 3.4296)
	insertListing(t, pool, agentID, "no-price", "sale", 9000000, 3, 6.4282, 3.4210)
	insertListing(t, pool, agentID, "no-bed",   "sale", 2000000, 1, 6.4283, 3.4220)

	url := "/api/v1/listings/search?type=sale&min_price=1000000&max_price=5000000&bedrooms=3"
	w := doRequest(t, r, http.MethodGet, url, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
	assert.Equal(t, "match", data[0].(map[string]interface{})["title"])
}

func TestSearch_InvalidType(t *testing.T) {
	pool := testPool(t)
	r := testRouter(pool)

	w := doRequest(t, r, http.MethodGet, "/api/v1/listings/search?type=unknown", nil)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
