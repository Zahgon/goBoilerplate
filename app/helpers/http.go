package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var jsonContentType = []string{"application/json"}

// JSONRender renders a value as JSON on the wire: the content type carries no
// charset parameter and the encoded value is terminated by a newline.
type JSONRender struct {
	Data interface{}
}

// Render writes the JSON encoding of the value to the response.
func (r JSONRender) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	return json.NewEncoder(w).Encode(r.Data)
}

// WriteContentType sets the JSON content type unless one is already set.
func (r JSONRender) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if len(header["Content-Type"]) == 0 {
		header["Content-Type"] = jsonContentType
	}
}

// JSON helper to send a JSON response with the given status code
func JSON(c *gin.Context, code int, i interface{}) {
	c.Render(code, JSONRender{Data: i})
}

// ErrorJSON sends the router's error envelope for a status code. A HEAD
// request is answered with the status alone, so the response carries neither a
// body nor a content type.
func ErrorJSON(c *gin.Context, code int) {
	if c.Request.Method == http.MethodHead {
		c.Status(code)
		c.Writer.WriteHeaderNow()
		return
	}

	JSON(c, code, gin.H{"message": http.StatusText(code)})
}

// HTTPError carries a status code alongside a message.
type HTTPError struct {
	Code    int         `json:"-"`
	Message interface{} `json:"message"`
}

// Error makes HTTPError an error.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("code=%d, message=%v", e.Code, e.Message)
}

// NewHTTPError helper to build an HTTPError
func NewHTTPError(code int, message interface{}) *HTTPError {
	return &HTTPError{Code: code, Message: message}
}
