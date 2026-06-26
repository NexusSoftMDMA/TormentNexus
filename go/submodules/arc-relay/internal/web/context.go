package web

import (
	"context"

	"github.com/comma-compliance/arc-relay/internal/store"
)

type contextKey string

const (
	userKey      contextKey = "user"
	sessionIDKey contextKey = "sessionID"
)

func setUser(ctx context.Context, user *store.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func getUser(r interface{ Context() context.Context }) *store.User {
	u, _ := r.Context().Value(userKey).(*store.User)
	return u
}

func setSessionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, sessionIDKey, id)
}

func getSessionID(ctx context.Context) string {
	s, _ := ctx.Value(sessionIDKey).(string)
	return s
}
