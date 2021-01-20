package middleware

import (
	goErrors "errors"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.unanet.io/devops/go/pkg/errors"
)

func TestGetReqID(t *testing.T) {
	var err error
	err = errors.RestError{
		Code:    400,
		Message: "blah",
	}

	if err, ok := err.(error); ok {
		var restError errors.RestError
		if goErrors.As(err, &restError) {
			require.Equal(t, "blah", restError.Message)
		}

		require.NotNil(t, restError)

	}
}
