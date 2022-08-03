package cm

import (
	"context"
	"fmt"
)

type ctxKeyMessaging int

const ContextKeyID ctxKeyMessaging = 0

type M struct {
	messages []string
}

func (m *M) Message(format string, a ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf(format, a...))
}

func (m *M) Messages() []string {
	return m.messages
}

type Messenger interface {
	Message(format string, a ...interface{})
	Messages() []string
}

func NewMessenger() Messenger {
	return &M{}
}

func getMessenger(ctx context.Context) Messenger {
	if ctx == nil {
		return nil
	}

	if m, ok := ctx.Value(ContextKeyID).(Messenger); ok {
		return m
	}

	return nil
}

func Messages(ctx context.Context) []string {
	if x := getMessenger(ctx); x == nil {
		return []string{}
	} else {
		return x.Messages()
	}
}

func Message(ctx context.Context, format string, a ...interface{}) {
	if x := getMessenger(ctx); x == nil {
		return
	} else {
		x.Message(format, a)
	}
}
