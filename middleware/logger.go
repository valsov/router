package middleware

import (
	"net/http"
	"time"
)

type RequestLogger interface {
	LogRequest(req *http.Request, elapsed time.Duration)
}

func LoggerMiddleware(logger RequestLogger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			defer func() {
				logger.LogRequest(r, time.Since(start))
			}()
			next.ServeHTTP(w, r)
		})
	}
}
