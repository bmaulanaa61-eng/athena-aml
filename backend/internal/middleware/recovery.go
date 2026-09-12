package middleware

import (
	"fmt"
	"net/http"

	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/errors"
	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/response"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				err, ok := rvr.(error)
				if !ok {
					err = fmt.Errorf("%v", rvr)
				}
				response.Error(w, r, errors.Internal(err, "Terjadi kepanikan tak terduga pada server."))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
