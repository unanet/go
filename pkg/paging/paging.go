package paging

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	uuid "github.com/satori/go.uuid"
)

type ctxKeyPaging int

const ContextKeyID ctxKeyPaging = 0

func GetParameters(ctx context.Context) *Parameters {
	if ctx == nil {
		return nil
	}

	if m, ok := ctx.Value(ContextKeyID).(*Parameters); ok {
		return m
	}

	return nil
}

type Cursor struct {
	IntID     int       `json:"int_id"`
	CreatedAt time.Time `json:"created_at"`
	UUID      uuid.UUID `json:"uuid"`
}

func (c Cursor) String() string {
	b, err := json.Marshal(&c)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

type Parameters struct {
	Limit  int     `json:"limit"`
	Cursor *Cursor `json:"cursor"`
}
