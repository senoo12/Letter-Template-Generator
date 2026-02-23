package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BaseResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, BaseResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, status int, message string, err interface{}) {
	c.JSON(status, BaseResponse{
		Success: false,
		Message: message,
		Error:   err,
	})
}

func OK(c *gin.Context, data interface{}) {
	Success(c, http.StatusOK, "success", data)
}

func Created(c *gin.Context, data interface{}) {
	Success(c, http.StatusCreated, "created", data)
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message, nil)
}

func InternalError(c *gin.Context, err interface{}) {
	Error(c, http.StatusInternalServerError, "internal server error", err)
}