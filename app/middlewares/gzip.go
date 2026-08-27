package middlewares

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// gzipResponseWriter compresses the body while leaving the response headers in
// the state the framework expects: Content-Encoding is announced only once a
// body is actually produced, and Content-Length is dropped because the encoded
// length is not known up front.
type gzipResponseWriter struct {
	gin.ResponseWriter
	writer    *gzip.Writer
	wroteBody bool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	header := w.Header()
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", http.DetectContentType(b))
	}
	header.Set("Content-Encoding", "gzip")
	header.Del("Content-Length")
	w.wroteBody = true
	return w.writer.Write(b)
}

func (w *gzipResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// Gzip Middleware
func Gzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.FullPath(), "/login") {
			return
		}

		c.Writer.Header().Add("Vary", "Accept-Encoding")

		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			return
		}

		gz, err := gzip.NewWriterLevel(c.Writer, gzip.DefaultCompression)
		if err != nil {
			return
		}

		original := c.Writer
		wrapped := &gzipResponseWriter{ResponseWriter: original, writer: gz}
		c.Writer = wrapped

		c.Next()

		if wrapped.wroteBody {
			gz.Close()
		} else {
			original.Header().Del("Content-Encoding")
		}
		c.Writer = original
	}
}
