package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

type requestIDKey string

const (
	RequestIDCtxKey requestIDKey = "request_id"
	HeaderRequestID              = "X-Request-ID"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(HeaderRequestID)
		if reqID == "" {
			reqID = generateRandomID()
		}

		w.Header().Set(HeaderRequestID, reqID)

		ctx := context.WithValue(r.Context(), RequestIDCtxKey, reqID)
		ctx = context.WithValue(ctx, "request_id", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRandomID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "req_fallback_00000000"
	}
	return fmt.Sprintf("req_%x", b)
}
