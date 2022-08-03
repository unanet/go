package paging

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetParameters(t *testing.T) {
	pp := Parameters{
		Limit:  10,
		Cursor: nil,
	}
	ctx := context.WithValue(context.TODO(), ContextKeyID, &pp)

	if _, ok := ctx.Value(ContextKeyID).(*Parameters); ok {
		require.True(t, ok)
	}
}
