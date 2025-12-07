package middleware

import (
	"net/http"

	aegis "github.com/dovakiin0/goaegis-core/aegis/core"
)

// Simple middleware for net/http that uses a subject extractor provided by the user.
func Require(a *aegis.Aegis, subjectExtractor func(r *http.Request) string, resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subject := subjectExtractor(r)
			ok, err := a.IsAllowed(subject, resource, action)
			if err != nil || !ok {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
