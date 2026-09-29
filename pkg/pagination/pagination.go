package pagination

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

type Params struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

func FromContext(c *gin.Context) Params {
	page := parseIntQuery(c, "page", DefaultPage)
	perPage := parseIntQuery(c, "per_page", DefaultPerPage)

	if page < 1 {
		page = DefaultPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	return Params{Page: page, PerPage: perPage}
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func (p Params) Limit() int {
	return p.PerPage
}

func NewMeta(params Params, total int64) Meta {
	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	if totalPages < 1 {
		totalPages = 1
	}
	return Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalItems: total,
		TotalPages: totalPages,
	}
}

func parseIntQuery(c *gin.Context, key string, defaultVal int) int {
	raw := c.Query(key)
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultVal
	}
	return v
}
