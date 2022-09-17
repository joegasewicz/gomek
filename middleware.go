package gomek

import (
	"fmt"
	"github.com/joegasewicz/status-writer"
	"net/http"
	"strings"
	"time"
)

// Logging adds logging for each request
func Logging(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Log response
		duration := time.Duration(time.Now().Sub(start)) * time.Nanosecond

		// Set status
		sw := status_writer.New(w)
		next.ServeHTTP(sw, r)
		statusCode := sw.Status
		msg := fmt.Sprintf("[INFO] %s %s %ds Status: %d\n", r.Method, r.RequestURI, duration, statusCode)
		out := PrintWithColor(msg, BLUE)
		fmt.Printf(out)
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

func allowRoute(routes [][]string, currentRoute string, reqMethod string) bool {
	for _, r := range routes {
		route, method := r[0], r[1]
		splitPath := strings.Split(route, "*")
		if len(splitPath) == 2 {
			// matchEndPath is the last path segment before the '/*'
			matchEndPath := splitPath[0]
			// matchEndPath should match the first part of the current route
			currentRouteMatch := strings.Split(currentRoute, matchEndPath)
			if currentRouteMatch[0] == "" && method == reqMethod {
				return true
			}
		} else {
			if route == currentRoute && method == reqMethod {
				return true
			}
		}
	}
	return false
}

// Authorize If you use the `gomek.Authorize` middleware, all your routes will need to pass authorization
// via the callback function passed to `gomek.Authorize`. To whitelist routes, pass a list of string
// pairs, representing the path and the request method, respectively.
//
//				var whiteList = [][]string{
//						{
//							"/", "GET",
//						},
//						{
//							"/login", "GET",
//						},
//				}
//
//	The `gomek.Authorize` middleware function require 2 arguments, your `[][]string` of path / request methods
//	and a callback function to test your auth strategy (e.g. session  or JWT).
//
//				app.Use(gomek.Authorize(whiteList, func() {
//	   				// Add your session / JWT test logic here.
//					// Return here if your auth logic fails.
//				}))
func Authorize(whiteList [][]string, callback func() bool) func(next http.Handler) http.HandlerFunc {
	inner := func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ok := allowRoute(whiteList, r.RequestURI, r.Method); !ok {
				// This route is not whitelisted so perform test from callback
				if ok := callback(); !ok {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
	return inner
}
