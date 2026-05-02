package middleware

import (
	"net/http"
	"strings"
)

const (
	jsonBodyLimit      int64 = 1 << 20
	multipartBodyLimit int64 = 60 << 20
)

func LimitarBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := jsonBodyLimit
		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		if strings.HasPrefix(contentType, "multipart/form-data") {
			limit = multipartBodyLimit
		}

		r.Body = http.MaxBytesReader(w, r.Body, limit)
		next.ServeHTTP(w, r)
	})
}
