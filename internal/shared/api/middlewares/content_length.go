package middlewares

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
)

var methodsToCheck = map[string]struct{}{
	http.MethodPost: {},
	http.MethodPut:  {},
}

func ContentLengthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := methodsToCheck[r.Method]; ok {
				// immediately return error if content length exceeds cfg maxContentLength size
				if r.ContentLength > int64(cfg.Server.MaxContentLength) {
					http.Error(w, "Request Entity Too Large", http.StatusRequestEntityTooLarge)
					return
				}
				// limit the size of the request body
				r.Body = http.MaxBytesReader(w, r.Body, int64(cfg.Server.MaxContentLength))
			}
			next.ServeHTTP(w, r)
		})
	}
}
