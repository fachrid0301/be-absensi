package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// JSONSuccess kirim response sukses format standar API.
func JSONSuccess(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// JSONError kirim response gagal format standar API.
func JSONError(c *gin.Context, status int, message string, errors interface{}) {
	c.JSON(status, APIResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// JSONValidationError kirim error validasi input (HTTP 400).
func JSONValidationError(c *gin.Context, errors interface{}) {
	JSONError(c, http.StatusBadRequest, "validasi gagal", errors)
}
