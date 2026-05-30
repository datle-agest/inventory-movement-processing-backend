package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type responseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
}

func ResponseTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			status:         http.StatusOK,
		}
		c.Writer = rw

		c.Next()

		elapsed := time.Since(start).Milliseconds()
		fmt.Println(">>> ResponseTime middleware ran:", elapsed, "ms")

		// Set header trước khi thực sự write response
		rw.ResponseWriter.Header().Set("X-Response-Time", strconv.FormatInt(elapsed, 10)+"ms")
		rw.ResponseWriter.WriteHeader(rw.status)
		rw.ResponseWriter.Write(rw.body.Bytes())
	}
}
