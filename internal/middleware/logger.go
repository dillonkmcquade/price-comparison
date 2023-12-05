package middleware

import (
	"log"
	"net/http"
)

// The current logger format is:
//
//	log.Printf("%s - %s - [%s] %s", r.RemoteAddr, r.Proto, r.Method, r.URL)
func Logger(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s - %s - [%s] %s", r.RemoteAddr, r.Proto, r.Method, r.URL)
		next.ServeHTTP(w, r)
	}
}
