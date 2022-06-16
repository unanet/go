package middleware

import (
	"context"
	"net/http"

	"github.com/unanet/go/pkg/cm"
)

func Messaging(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, cm.MessagingKeyID, cm.NewMessenger())
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
