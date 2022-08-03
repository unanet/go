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

func Paging(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		pp, err := pageParameters(r)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		ctx = context.WithValue(ctx, paging.ContextKeyID, &pp)
		defer func() {
			w.Header().Add(
				"x-paging-cursor",
				pp.Cursor.String(),
			)
		}()
		next.ServeHTTP(w, r.WithContext(ctx))

	}
	return http.HandlerFunc(fn)
}

func pageParameters(r *http.Request) (*paging.Parameters, error) {
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		return &paging.Parameters{
			Limit:  100,
			Cursor: nil,
		}, nil
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return nil, errors.BadRequest("limit query parameter must be an int")
	}

	cursor := r.URL.Query().Get("cursor")
	if cursor == "" {
		return &paging.Parameters{
			Limit:  limitInt,
			Cursor: nil,
		}, nil
	} else {
		bcursor, err := base64.StdEncoding.DecodeString(cursor)
		if err != nil {
			return nil, errors.BadRequest("invalid cursor query parameter")
		}
		dcursor := paging.Cursor{}
		err = json.Unmarshal(bcursor, &dcursor)
		if err != nil {
			return nil, errors.BadRequest("invalid cursor query parameter")
		}

		return &paging.Parameters{
			Limit:  limitInt,
			Cursor: &dcursor,
		}, nil
	}

}
