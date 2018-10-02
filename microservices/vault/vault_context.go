package vault

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const ContextKey = "context"

type Context struct {
	Log *logrus.Logger
}

type contextHandler struct {
	context *Context
	handler http.Handler
}

func NewContextHandler(ctx *Context, next http.Handler) http.Handler {
	return contextHandler{ctx, next}
}

func (c contextHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := context.WithValue(request.Context(), ContextKey, c.context)

	request = request.WithContext(ctx)
	c.handler.ServeHTTP(writer, request)
}

func GetContext(request *http.Request) (*Context, error) {
	c := request.Context().Value(ContextKey)
	if value, ok := c.(*Context); ok {
		return value, nil
	}
	return nil, errors.Errorf("context error")
}
