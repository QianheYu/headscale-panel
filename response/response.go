package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response to frontend
func Response(c *gin.Context, httpStatus int, code int, data interface{}, message string) {
	c.JSON(httpStatus, gin.H{"code": code, "data": data, "message": message})
}

// Response to frontend - Success
func Success(c *gin.Context, data interface{}, message string) {
	Response(c, http.StatusOK, 200, data, message)
}

// Response to frontend - Fail
func Fail(c *gin.Context, data interface{}, message string) {
	Response(c, http.StatusBadRequest, 400, data, message)
}
