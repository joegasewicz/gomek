package gomek

import (
	"fmt"
	"net/http"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func Logging(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := statusWriter{ResponseWriter: w}
		// Log response
		duration := time.Duration(time.Now().Sub(start)) * time.Nanosecond
		var status int
		if sw.status == 0 {
			status = 200
		} else {
			status = sw.status
		}
		fmt.Printf("[INFO] %s %s %ds Status: %d\n", r.Method, r.RequestURI, duration, status)
		next.ServeHTTP(&sw, r)
	})
}
