package iq

import (
	"context"
	"time"
)

type MessageAttribute struct {
	Type  string `json:"Type"`
	Value string `json:"Value"`
}

type NotificationMessage struct {
	Type             string                      `json:"Type"`
	ID               string                      `json:"MessageId"`
	TopicArn         string                      `json:"TopicArn"`
	Body             string                      `json:"Message"`
	Timestamp        time.Time                   `json:"Timestamp"`
	SignatureVersion string                      `json:"SignatureVersion"`
	Signature        string                      `json:"Signature"`
	SigningCertURL   string                      `json:"SigningCertURL"`
	UnsubscribeURL   string                      `json:"UnsubscribeURL"`
	Attributes       map[string]MessageAttribute `json:"Attributes"`
}

type M struct {
	Notification  NotificationMessage
	ReceiptHandle string
	Body          map[string]interface{}
	ID            string
}

type mContext struct {
	M
	ctx context.Context
}
