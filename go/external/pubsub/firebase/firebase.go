package firebase

import (
	"context"
	"errors"

	"firebase.google.com/go/messaging"
	"gocloud.dev/gcerrors"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/batcher"
	"gocloud.dev/pubsub/driver"
)

var sendBatchOpts = &batcher.Options{
	MaxBatchSize: 500,
	MaxHandlers:  50, // TODO(taekyeom) maybe tuned..
}

type TopicOptions struct {
	DryRun bool

	BacherOptions *batcher.Options
}

type fcmTopic struct {
	client *messaging.Client
	opts   *TopicOptions
}

func OpenFCMTopic(ctx context.Context, client *messaging.Client, opts *TopicOptions) *pubsub.Topic {
	bo := sendBatchOpts
	if opts != nil && opts.BacherOptions != nil {
		bo = bo.NewMergedOptions(opts.BacherOptions)
	}
	return pubsub.NewTopic(openFCMTopic(ctx, client, opts), bo)
}

func openFCMTopic(ctx context.Context, client *messaging.Client, opts *TopicOptions) driver.Topic {
	if opts == nil {
		opts = &TopicOptions{}
	}
	return &fcmTopic{
		client: client,
		opts:   opts,
	}
}

func (t *fcmTopic) SendBatch(ctx context.Context, dms []*driver.Message) error {
	entries := make([]*messaging.Message, 0, len(dms))
	for _, dm := range dms {
		entry := &messaging.Message{}
		if err := entry.UnmarshalJSON(dm.Body); err != nil {
			return err
		}
		if dm.BeforeSend != nil {
			asFunc := func(i interface{}) bool {
				if p, ok := i.(**messaging.Message); ok {
					*p = entry
					return true
				}
				return false
			}
			if err := dm.BeforeSend(asFunc); err != nil {
				return err
			}
		}
		entries = append(entries, entry)
	}
	var err error
	var resp *messaging.BatchResponse
	if t.opts.DryRun {
		resp, err = t.client.SendAllDryRun(ctx, entries)
	} else {
		resp, err = t.client.SendAll(ctx, entries)
	}
	if err != nil {
		return err
	}

	// TODO(taekyeom) when failure count is non zero... handle these..

	if resp.SuccessCount == len(dms) {
		for n, dm := range dms {
			if dm.AfterSend != nil {
				asFunc := func(i interface{}) bool {
					if p, ok := i.(**messaging.SendResponse); ok {
						*p = resp.Responses[n]
						return true
					}
					return false
				}
				if err := dm.AfterSend(asFunc); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *fcmTopic) IsRetryable(error) bool {
	// The client handles retries.
	return false
}

func (t *fcmTopic) As(i interface{}) bool {
	c, ok := i.(**messaging.Client)
	if !ok {
		return false
	}
	*c = t.client
	return true
}

func (t *fcmTopic) ErrorAs(err error, i interface{}) bool {
	return errors.As(err, i)
}

func (t *fcmTopic) ErrorCode(err error) gcerrors.ErrorCode {
	if err != nil {
		return gcerrors.OK
	}
	// TODO (taekyeom) sophisticated error code
	return gcerrors.Unknown
}

func (t *fcmTopic) Close() error {
	return nil
}
