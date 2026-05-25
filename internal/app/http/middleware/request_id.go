package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

const HeaderRequestID = "X-Request-ID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderRequestID)
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set(HeaderRequestID, id)
		next.ServeHTTP(w, r)
	})
}
