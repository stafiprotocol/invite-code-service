package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	MaxPageSize     = 50
	DefaultPageSize = 10
)

type Rsp struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func Ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "80000",
		"message": "success",
		"data":    data,
	})
}

func Err(c *gin.Context, status, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"status":  status,
		"message": msg,
		"data":    struct{}{},
	})
}
