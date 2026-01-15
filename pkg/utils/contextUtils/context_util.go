package contextutils

import (
	"context"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func SetRequestIDInContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}
