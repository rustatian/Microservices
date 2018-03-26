package old

import (
	"context"
	"github.com/nats-io/go-nats"
)

type DecodeRequestFunc func(_ context.Context, msg *nats.Msg) (interface{}, error)

type EncodeRequestFunc func(_ context.Context, msg interface{}) ([]byte, error)

type EncodeResponseFunc func(_ context.Context, response interface{}) ([]byte, error)

type DecodeResponseFunc func(_ context.Context, msg *nats.Msg) (interface{}, error)
