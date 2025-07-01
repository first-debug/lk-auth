package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logging(log *slog.Logger) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			defer func() {
				log.Info(
					r.URL.Path,
					"duration", time.Since(start).String(),
				)
			}()
			f(w, r)
		}
	}
}
