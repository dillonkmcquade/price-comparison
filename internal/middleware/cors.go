package middleware

import (
	"net/http"
)

// Sets cors headers
func Cors(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		next.ServeHTTP(w, r)
	}
}
