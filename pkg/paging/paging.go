package paging

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
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

	// this use case gets hit when middleware isn't being used/set unit tests as well
	p := NewParameters(1000, nil, nil)
	return &p
}

type Cursor struct {
	IntID     *int       `json:"int_id"`
	CreatedAt *time.Time `json:"created_at"`
	UUID      *uuid.UUID `json:"uuid"`
}

func (c Cursor) String() string {
	b, err := json.Marshal(&c)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func NewParameters(limit uint64, cursor *Cursor, w http.ResponseWriter) Parameters {
	return Parameters{
		Limit:  limit,
		Cursor: cursor,
		w:      w,
	}
}

type Parameters struct {
	Limit  uint64  `json:"limit"`
	Cursor *Cursor `json:"cursor"`
	w      http.ResponseWriter
}

func (p Parameters) SetIntCursor(id int) {
	p.w.Header().Add(
		"x-paging-cursor",
		Cursor{IntID: &id}.String(),
	)
}

func (p Parameters) SetUUIDCursor(uuid uuid.UUID, createdAt time.Time) {
	p.w.Header().Add(
		"x-paging-cursor",
		Cursor{CreatedAt: &createdAt, UUID: &uuid}.String(),
	)
}
