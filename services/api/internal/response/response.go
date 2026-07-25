package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope is the standard API response shape.
type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    any         `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

// ErrorBody carries machine-readable error details.
type ErrorBody struct {
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// JSON writes a successful envelope.
func JSON(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error writes a failed envelope.
func Error(c *gin.Context, status int, message, code, details string) {
	c.JSON(status, Envelope{
		Success: false,
		Message: message,
		Error: &ErrorBody{
			Code:    code,
			Details: details,
		},
	})
}

// OK is a convenience for HTTP 200 success responses.
func OK(c *gin.Context, message string, data any) {
	JSON(c, http.StatusOK, message, data)
}
