package clients

import (
	"context"

	"google.golang.org/grpc/metadata"
)

var forwardedIdentityKeys = []string{
	"x-user-id",
	"x-user-role",
	"x-user-email",
	"x-user-email-verified",
}

func withForwardedIdentity(ctx context.Context) context.Context {
	incoming, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	pairs := make([]string, 0, len(forwardedIdentityKeys)*2)
	for _, key := range forwardedIdentityKeys {
		values := incoming.Get(key)
		if len(values) == 0 {
			continue
		}
		pairs = append(pairs, key, values[0])
	}
	if len(pairs) == 0 {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, pairs...)
}
