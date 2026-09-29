package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Data interface{} `json:"data"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination interface{} `json:"pagination"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func OKPaginated(c *gin.Context, data interface{}, pagination interface{}) {
	c.JSON(http.StatusOK, PaginatedResponse{Data: data, Pagination: pagination})
}
