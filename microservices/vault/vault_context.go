package vault

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
)

const ContextKey = "context"

type VaultContext struct {
	Log *logrus.Logger
}

type contextHandler struct {
	context *VaultContext
	handler http.Handler
}

func NewContextHandler(ctx *VaultContext, next http.Handler) http.Handler {
	return contextHandler{ctx, next}
}

func (c contextHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := context.WithValue(request.Context(), ContextKey, c.context)

	request = request.WithContext(ctx)
	c.handler.ServeHTTP(writer, request)
}

func GetContext(request *http.Request) (*VaultContext, error) {
	c := request.Context().Value(ContextKey)
	if value, ok := c.(*VaultContext); ok {
		return value, nil
	}
	return nil, errors.Errorf("context error")
}
