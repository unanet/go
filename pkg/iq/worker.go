package iq

import (
	"context"
	gjson "encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"

	"github.com/unanet/go/pkg/errors"
	"github.com/unanet/go/pkg/log"
)

// HandlerFunc is used to define the Handler that is run on for each message
type HandlerFunc func(ctx context.Context, msg *M) error

// HandleMessage wraps a function for handling sqs messages
func (f HandlerFunc) HandleMessage(ctx context.Context, msg *M) error {
	return f(ctx, msg)
}

type Handler interface {
	HandleMessage(ctx context.Context, msg *M) error
}

const (
	MessageAttributeReqID string = "x_req_id"
)

type Config struct {
	Prefix                 string   `split_words:"true" required:"true"`
	TopicArns              []string `split_words:"true" required:"true"`
	MaxNumberOfMessages    int64    `split_words:"true" default:"10"`
	WaitTimeSecond         int64    `split_words:"true" default:"20"`
	VisibilityTimeout      int64    `split_words:"true" default:"3600"`
	DeliveryDelay          int64    `split_words:"true" default:"0"`
	MessageRetentionPeriod int64    `split_words:"true" default:"3600"`
	HandlerTimeout         int64    `split_words:"true" default:"60"`
}

type InstanceQ struct {
	log           *zap.Logger
	name          string
	ctx           context.Context
	cancel        context.CancelFunc
	done          chan bool
	sns           *sns.SNS
	sqs           *sqs.SQS
	c             *Config
	qurl          string
	qarn          string
	subscriptions []string
}

func NewInstanceQ(instanceName string, sess *session.Session, c *Config) (*InstanceQ, error) {
	w := InstanceQ{
		name: instanceName,
		log:  log.Logger.With(zap.String("worker", instanceName)),
		c:    c,
		sqs:  sqs.New(sess),
		sns:  sns.New(sess),
		done: make(chan bool),
	}

	return &w, nil
}

func getInstanceID(instanceName string) string {
	splitInstanceName := strings.Split(instanceName, "-")
	if len(splitInstanceName) > 2 {
		return splitInstanceName[len(splitInstanceName)-2]
	} else {
		return "0000000000"
	}
}

func (q *InstanceQ) createQ() error {
	instanceID := getInstanceID(q.name)

	result, err := q.sqs.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(fmt.Sprintf("%s_srv-%s", q.c.Prefix, q.name)),
		Attributes: map[string]*string{
			"DelaySeconds":           aws.String(fmt.Sprint(q.c.DeliveryDelay)),
			"VisibilityTimeout":      aws.String(fmt.Sprint(q.c.VisibilityTimeout)),
			"MessageRetentionPeriod": aws.String(fmt.Sprint(q.c.MessageRetentionPeriod)),
		},
		Tags: map[string]*string{
			"Prefix":     aws.String(q.c.Prefix),
			"InstanceID": aws.String(instanceID),
		},
	})
	if err != nil {
		return err
	}

	qAttrs, err := q.sqs.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		AttributeNames: aws.StringSlice([]string{"QueueArn"}),
		QueueUrl:       result.QueueUrl,
	})
	if err != nil {
		return err
	}

	qarn := *qAttrs.Attributes["QueueArn"]

	var subscriptions []string
	var policies []interface{}

	log.Logger.Debug("subscribe Q ARns", zap.String("qarn", qarn), zap.Strings("topic_arns", q.c.TopicArns))

	for _, x := range q.c.TopicArns {
		log.Logger.Debug("subscription to topic ARns", zap.String("arn", x))

		r, err := q.sns.Subscribe(&sns.SubscribeInput{
			Endpoint: aws.String(qarn),
			Protocol: aws.String("sqs"),
			TopicArn: aws.String(x),
		})
		if err != nil {
			log.Logger.Error("failed to subscribe to topic", zap.String("topic", x), zap.String("iq", qarn), zap.Error(err))
			return err
		} else {
			subscriptions = append(subscriptions, *r.SubscriptionArn)
			policies = append(policies, getSqsPolicy(qarn, x))
		}
	}

	log.Logger.Debug("queue policies", zap.Any("policies", policies))

	b, err := gjson.Marshal(map[string]interface{}{
		"Statement": policies,
	})
	if err != nil {
		log.Logger.Error("failed to marshal sqs policies", zap.Error(err))
	}

	policy := string(b)

	_, err = q.sqs.SetQueueAttributes(&sqs.SetQueueAttributesInput{
		Attributes: map[string]*string{
			"Policy": aws.String(policy),
		},
		QueueUrl: result.QueueUrl,
	})
	if err != nil {
		log.Logger.Error("failed to set sqs policy", zap.Error(err), zap.String("policy", policy))
	}

	q.subscriptions = subscriptions
	q.qurl = *result.QueueUrl
	q.qarn = qarn

	return nil
}

func (q *InstanceQ) Start(h Handler) {
	ctx, cancel := context.WithCancel(context.Background())
	q.ctx = ctx
	q.cancel = cancel

	go func() {
		q.log.Info("instance queue worker started")
		for {
			select {
			case <-q.ctx.Done():
				q.log.Info("instance queue worker stopped")
				close(q.done)
				return
			default:
				ctx := context.Background()
				m, err := q.receive(ctx)
				if err != nil {
					q.log.Panic("error receiving message from queue", zap.Error(err))
				}
				if len(m) == 0 {
					continue
				}
				q.run(h, m)
			}
		}
	}()
}

func (q *InstanceQ) cleanup() {
	if q.sns != nil {
		for _, x := range q.subscriptions {
			q.log.Info("unsubscribing from SNS Topic", zap.String("subscription", x))
			if _, err := q.sns.Unsubscribe(&sns.UnsubscribeInput{
				SubscriptionArn: aws.String(x),
			}); err != nil {
				q.log.Error("error unsubscribing from SNS Topic", zap.Error(err), zap.String("subscription", x))
			}
		}
	}

	if q.sqs != nil {
		q.log.Info("deleting SQS Queue", zap.String("name", q.qurl))
		_, err := q.sqs.DeleteQueue(&sqs.DeleteQueueInput{
			QueueUrl: aws.String(q.qurl),
		})
		if err != nil {
			q.log.Error("error deleting SQS Topic", zap.Error(err), zap.String("qurl", q.qurl))
		}
	}
}

func (q *InstanceQ) Stop() {
	q.cancel()
	<-q.done
	q.cleanup()
}

func (q *InstanceQ) run(h Handler, mCtx []*mContext) {
	numMessages := len(mCtx)
	var wg sync.WaitGroup
	wg.Add(numMessages)
	for _, mc := range mCtx {
		go func(m *mContext) {
			ctx, cancel := context.WithTimeout(m.ctx, time.Duration(q.c.HandlerTimeout)*time.Second)
			defer cancel()
			defer wg.Done()
			if err := h.HandleMessage(ctx, &m.M); err != nil {
				q.log.Error("error handling message", zap.Error(err))
			} else {
				err = q.deleteMessage(m.ctx, &m.M)
				if err != nil {
					q.log.Error("error deleting message", zap.Error(err))
				}
			}
		}(mc)
	}
	wg.Wait()
}

func (q *InstanceQ) logWith(ctx context.Context) *zap.Logger {
	return q.log.With(zap.String("req_id", log.GetReqID(ctx)))
}

func (q *InstanceQ) receive(ctx context.Context) ([]*mContext, error) {
	awsM := sqs.ReceiveMessageInput{
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            aws.String(q.qurl),
		MaxNumberOfMessages: aws.Int64(q.c.MaxNumberOfMessages),
		VisibilityTimeout:   aws.Int64(q.c.VisibilityTimeout),
		WaitTimeSeconds:     aws.Int64(q.c.WaitTimeSecond),
	}
	result, err := q.sqs.ReceiveMessage(&awsM)
	if err != nil {
		if strings.HasPrefix(err.Error(), "RequestCanceled") {
			return nil, nil
		}
		return nil, errors.Wrap(err)
	}

	var returnMs []*mContext
	for _, x := range result.Messages {
		var n NotificationMessage
		err = gjson.Unmarshal([]byte(*x.Body), &n)
		if err != nil {
			log.Logger.Error("failed to unmarshal notification message", zap.Error(err))
		}

		body := make(map[string]interface{})
		err = gjson.Unmarshal([]byte(n.Body), &body)
		if err != nil {
			log.Logger.Error("failed to unmarshal notification body", zap.Error(err))
		}
		m := M{
			ID:            *x.MessageId,
			Notification:  n,
			Body:          body,
			ReceiptHandle: *x.ReceiptHandle,
		}
		var mctx context.Context
		if val, ok := n.Attributes[MessageAttributeReqID]; ok {
			mctx = context.WithValue(ctx, log.RequestIDKey, val.Value)
		} else {
			mctx = context.WithValue(ctx, log.RequestIDKey, "00000000000000000000000000000000")
		}
		returnMs = append(returnMs, &mContext{
			M:   m,
			ctx: mctx,
		})
		q.logWith(mctx).Info("notification message received",
			zap.Any("id", m.ID),
		)
	}

	return returnMs, nil
}

func (q *InstanceQ) deleteMessage(ctx context.Context, m *M) error {
	now := time.Now()
	_, err := q.sqs.DeleteMessageWithContext(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(q.qurl),
		ReceiptHandle: aws.String(m.ReceiptHandle),
	})
	if err != nil {
		return errors.Wrap(err)
	}
	q.logWith(ctx).Info("notification message deleted",
		zap.Float64("elapsed_ms", float64(time.Since(now).Nanoseconds())/1000000.0),
		zap.Any("id", m.ID),
	)
	return nil
}

func GetLogger(ctx context.Context) *zap.Logger {
	reqID := log.GetReqID(ctx)
	if len(reqID) > 0 {
		return log.Logger.With(zap.String("req_id", reqID))
	} else {
		return log.Logger
	}
}

func getSqsPolicy(qArn, tArn string) map[string]interface{} {
	return map[string]interface{}{
		"Sid":    "Allow-SNS-SendMessage",
		"Effect": "Allow",
		"Principal": map[string]interface{}{
			"Service": "sns.amazonaws.com",
		},
		"Action": []interface{}{
			"sqs:SendMessage",
		},
		"Resource": qArn,
		"Condition": map[string]interface{}{
			"ArnEquals": map[string]interface{}{
				"aws:SourceArn": tArn,
			},
		},
	}
}
