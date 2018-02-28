package nats

import (
	"context"
	"encoding/base64"
	"strings"

	"google.golang.org/grpc/metadata"
)

const (
	binHdrSuffix = "-bin"
)

type ClientRequestFunc func(context.Context, *metadata.MD) context.Context

type ServerRequestFunc func(context.Context, metadata.MD) context.Context

type ServerResponseFunc func(ctx context.Context, header *metadata.MD, trailer *metadata.MD) context.Context

type ClientResponseFunc func(ctx context.Context, header metadata.MD, trailer metadata.MD) context.Context

func SetRequestHeader(key, val string) ClientRequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		key, val := EncodeKeyValue(key, val)
		(*md)[key] = append((*md)[key], val)
		return ctx
	}
}

func SetResponseHeader(key, val string) ServerResponseFunc {
	return func(ctx context.Context, md *metadata.MD, _ *metadata.MD) context.Context {
		key, val := EncodeKeyValue(key, val)
		(*md)[key] = append((*md)[key], val)
		return ctx
	}
}

func SetResponseTrailer(key, val string) ServerResponseFunc {
	return func(ctx context.Context, _ *metadata.MD, md *metadata.MD) context.Context {
		key, val := EncodeKeyValue(key, val)
		(*md)[key] = append((*md)[key], val)
		return ctx
	}
}

func EncodeKeyValue(key, val string) (string, string) {
	key = strings.ToLower(key)
	if strings.HasSuffix(key, binHdrSuffix) {
		v := base64.StdEncoding.EncodeToString([]byte(val))
		val = string(v)
	}
	return key, val
}
