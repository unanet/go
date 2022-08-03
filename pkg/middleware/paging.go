package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/render"

	"github.com/unanet/go/pkg/errors"
	"github.com/unanet/go/pkg/paging"
)

func Paging(defaultLimit uint64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			pp, err := pageParameters(r, w, defaultLimit)
			if err != nil {
				render.Respond(w, r, err)
				return
			}

			ctx = context.WithValue(ctx, paging.ContextKeyID, &pp)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

func pageParameters(r *http.Request, w http.ResponseWriter, defaultLimit uint64) (paging.Parameters, error) {
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		return paging.NewParameters(defaultLimit, nil, w), nil
	}

	limitInt, err := strconv.ParseUint(limit, 10, 64)
	if err != nil {
		return paging.Parameters{}, errors.BadRequest("limit query parameter must be an int")
	}

	cursor := r.URL.Query().Get("cursor")
	if cursor == "" {
		return paging.NewParameters(limitInt, nil, w), nil
	} else {
		bcursor, err := base64.StdEncoding.DecodeString(cursor)
		if err != nil {
			return paging.Parameters{}, errors.BadRequest("invalid cursor query parameter")
		}
		dcursor := paging.Cursor{}
		err = json.Unmarshal(bcursor, &dcursor)
		if err != nil {
			return paging.Parameters{}, errors.BadRequest("invalid cursor query parameter")
		}

		return paging.NewParameters(limitInt, &dcursor, w), nil
	}

}
