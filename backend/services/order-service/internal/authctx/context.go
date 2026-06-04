package authctx

import (
	"context"
	"crypto/subtle"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"google.golang.org/grpc/metadata"
)

const (
	maxActorIDLength   = 64
	maxRequestIDLength = 128
	maxSessionIDLength = 128
)

type Actor struct {
	UserID    string
	SellerID  string
	Roles     []string
	RequestID string
	SessionID string
}

type actorKey struct{}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, actor)
}

func ActorFromContext(ctx context.Context) (Actor, error) {
	actor, ok := ctx.Value(actorKey{}).(Actor)
	if !ok || strings.TrimSpace(actor.UserID) == "" {
		return Actor{}, domain.ErrUnauthenticated
	}
	return actor, nil
}

// AuthenticateIncoming verifies the caller credential before accepting
// identity metadata forwarded by a trusted gateway or internal service.
func AuthenticateIncoming(ctx context.Context, trustedToken string, fallbackRequestID string) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, domain.ErrUnauthenticated
	}
	presentedToken := first(md.Get("x-internal-token"))
	if presentedToken == "" || len(presentedToken) != len(trustedToken) ||
		subtle.ConstantTimeCompare([]byte(presentedToken), []byte(trustedToken)) != 1 {
		return ctx, domain.ErrUnauthenticated
	}
	userID := strings.TrimSpace(first(md.Get("x-user-id")))
	sellerID := strings.TrimSpace(first(md.Get("x-seller-id")))
	requestID := strings.TrimSpace(first(md.Get("x-request-id")))
	sessionID := strings.TrimSpace(first(md.Get("x-session-id")))
	if requestID == "" {
		requestID = fallbackRequestID
	}
	if userID == "" || len(userID) > maxActorIDLength ||
		len(sellerID) > maxActorIDLength ||
		len(requestID) > maxRequestIDLength || len(sessionID) > maxSessionIDLength {
		return ctx, domain.ErrUnauthenticated
	}
	roles := normalizeRoles(first(md.Get("x-roles")))
	return WithActor(ctx, Actor{
		UserID:    userID,
		SellerID:  sellerID,
		Roles:     roles,
		RequestID: requestID,
		SessionID: sessionID,
	}), nil
}

func HasRole(actor Actor, required ...string) bool {
	for _, role := range actor.Roles {
		for _, candidate := range required {
			if role == strings.ToLower(strings.TrimSpace(candidate)) {
				return true
			}
		}
	}
	return false
}

func normalizeRoles(value string) []string {
	seen := make(map[string]struct{})
	var roles []string
	for _, role := range strings.Split(value, ",") {
		role = strings.ToLower(strings.TrimSpace(role))
		if role == "" {
			continue
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		roles = append(roles, role)
	}
	return roles
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
