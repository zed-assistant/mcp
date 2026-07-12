package appctx

import "context"

type ctxKeyRequestId struct{}

func WithRequestId(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestId{}, requestId)
}

func GetRequestId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, ok := ctx.Value(ctxKeyRequestId{}).(string)
	if !ok {
		return ""
	}
	return id
}
