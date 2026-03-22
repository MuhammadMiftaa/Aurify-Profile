package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
)

// UnaryServerInterceptor returns a new unary server interceptor that extracts user metadata
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = extractUserMetadata(ctx)
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor that extracts user metadata
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := extractUserMetadata(ss.Context())
		wrapped := &wrappedStream{ServerStream: ss, ctx: ctx}
		return handler(srv, wrapped)
	}
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

func extractUserMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	if userIDs := md.Get("x-user-id"); len(userIDs) > 0 {
		ctx = context.WithValue(ctx, UserIDKey, userIDs[0])
	}

	return ctx
}

// UserIDFromContext extracts user ID from context
func UserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}
