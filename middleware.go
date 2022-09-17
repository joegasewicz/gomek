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

// Logging adds logging for each request
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
		msg := fmt.Sprintf("[INFO] %s %s %ds Status: %d\n", r.Method, r.RequestURI, duration, status)
		c := PrintWithColor(msg, BLUE)
		fmt.Printf(c)
		next.ServeHTTP(&sw, r)
	})
}

func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
}

// CORS basic development cors
func CORS(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
