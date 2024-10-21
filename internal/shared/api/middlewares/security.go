package middlewares

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/unrolled/secure"
)

const (
	swaggerPathPrefix = "/swagger/"
)

// SecurityHeadersMiddleware sets various security headers using the unrolled/secure package
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Default CSP
			defaultCSP := "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; object-src 'none'; frame-ancestors 'self'; form-action 'self'; block-all-mixed-content; base-uri 'self';"

			// CSP for /swagger/* path
			swaggerCSP := "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com https://stackpath.bootstrap.com; style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com https://stackpath.bootstrap.com; img-src 'self' data: https://cdnjs.cloudflare.com https://stackpath.bootstrap.com; font-src 'self' https://cdnjs.cloudflare.com https://stackpath.bootstrap.com; object-src 'none'; frame-ancestors 'self'; form-action 'self'; block-all-mixed-content; base-uri 'self';"

			// Choose the appropriate CSP based on the request path
			csp := defaultCSP
			if strings.HasPrefix(r.URL.Path, swaggerPathPrefix) {
				csp = swaggerCSP
			}

			// Create secure options with the chosen CSP
			sec := secure.New(secure.Options{
				FrameDeny:             true, // Equivalent to X-Frame-Options: DENY
				ContentTypeNosniff:    true, // Equivalent to X-Content-Type-Options: nosniff
				BrowserXssFilter:      true, // Equivalent to X-XSS-Protection: 1; mode=block
				ContentSecurityPolicy:  csp,
				ReferrerPolicy: "strict-origin-when-cross-origin", // Setting Referrer-Policy
			})

			// Apply the secure middleware
			err := sec.Process(w, r)
			if err != nil {
				// If there's an error, do not proceed further
				log.Error().Err(err).Msg("error while applying security headers")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}