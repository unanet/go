package middleware

import (
	"context"
	goErrors "errors"
	"net/http"

	"github.com/go-chi/render"
	"go.uber.org/zap"

	"github.com/unanet/go/pkg/errors"
)

func init() {
	render.Respond = func(w http.ResponseWriter, r *http.Request, v interface{}) {
		if err, ok := v.(error); ok {
			var restError errors.RestError
			if goErrors.As(err, &restError) {
				render.Status(r, restError.Code)
				LogFromRequest(r).Debug("Known Internal Server Error", zap.Error(err))
				render.DefaultResponder(w, r, restError)
				return
			}

			if goErrors.As(err, &context.Canceled) {
				contextCancelledError := errors.RestError{Code: 444, Message: "Context Cancelled", OriginalError: err}
				LogFromRequest(r).Warn("Context Cancelled", zap.Error(err))
				render.DefaultResponder(w, r, contextCancelledError)
				return
			}

			render.Status(r, 500)
			internalServerError := errors.RestError{Code: http.StatusInternalServerError, Message: "Internal Server Error", OriginalError: err}
			LogFromRequest(r).Error("Unknown Internal Server Error", zap.Error(err))
			render.DefaultResponder(w, r, internalServerError)
			return
		}
		render.DefaultResponder(w, r, v)
	}
}
