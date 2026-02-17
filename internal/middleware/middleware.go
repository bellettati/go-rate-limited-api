package middleware

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/bellettati/go-rate-limited-api/internal/limiter"
)

type StatusRecorder struct {
	http.ResponseWriter
	status int
}

func maskAPIKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}

	return key[:2] + "****" + key[len(key)-2:]
}

func (sr *StatusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func RateLimit(l limiter.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			recorder := &StatusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				http.Error(recorder, "missing API key", http.StatusUnauthorized)

				log.Printf(
					"method=%s path=%s apiKey=missing allowed=false status=%d duration=%s",
					r.Method,
					r.URL.Path,
					http.StatusUnauthorized,
					time.Since(start),
				)

				return
			}

			result := l.Allow(apiKey)

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

			if !result.Allowed {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)

				log.Printf(
					"method=%s path=%s apiKey=%s allowed=false status=%d remaining=%d duration=%s",
					r.Method,
					r.URL.Path,
					maskAPIKey(apiKey),
					http.StatusTooManyRequests,
					result.Remaining,
					time.Since(start),
				)

				return
			}

			log.Printf(
				"method=%s path=%s apiKey=%s allowed=false status=%d remaining=%d duration=%s",
				r.Method,
				r.URL.Path,
				maskAPIKey(apiKey),
				http.StatusOK,
				result.Remaining,
				time.Since(start),
			)

			next.ServeHTTP(w, r)
		})
	}
}
