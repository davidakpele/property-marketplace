package httpx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewBadRequest(msg string) *APIError {
	return &APIError{Code: http.StatusBadRequest, Message: msg}
}

func NewNotFound(msg string) *APIError {
	return &APIError{Code: http.StatusNotFound, Message: msg}
}

func NewUnprocessable(msg string) *APIError {
	return &APIError{Code: http.StatusUnprocessableEntity, Message: msg}
}

func NewInternal(msg string) *APIError {
	return &APIError{Code: http.StatusInternalServerError, Message: msg}
}

func NewConflict(msg string) *APIError {
	return &APIError{Code: http.StatusConflict, Message: msg}
}

func HandleError(c *gin.Context, err error) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		c.JSON(apiErr.Code, apiErr)
		return
	}
	c.JSON(http.StatusInternalServerError, &APIError{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
	})
}
