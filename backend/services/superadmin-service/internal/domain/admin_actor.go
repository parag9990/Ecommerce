package domain

import (
	"context"
	"strings"
)

type AdminActor struct {
	AdminID     string
	UserID      string
	Roles       []AdminRole
	SessionID   string
	RequestID   string
	IPHash      string
	MFAVerified bool
}

func (a AdminActor) ValidateForAdminRoute() error {
	if strings.TrimSpace(a.AdminID) == "" {
		return NewAdminContextMissing("admin_id is required for admin routes")
	}
	if strings.TrimSpace(a.UserID) == "" {
		return NewAdminContextMissing("user_id is required for admin routes")
	}
	if len(a.Roles) == 0 {
		return NewAdminContextMissing("at least one admin role is required")
	}
	if strings.TrimSpace(a.SessionID) == "" {
		return NewAdminContextMissing("session_id is required for admin routes")
	}
	return nil
}

func (a AdminActor) HasRole(role AdminRole) bool {
	for _, candidate := range a.Roles {
		if candidate == role {
			return true
		}
	}
	return false
}

func RolesToStrings(roles []AdminRole) []string {
	out := make([]string, 0, len(roles))
	for _, role := range roles {
		if role == "" {
			continue
		}
		out = append(out, string(role))
	}
	return out
}

func RolesFromStrings(values []string) []AdminRole {
	roles := make([]AdminRole, 0, len(values))
	seen := make(map[AdminRole]struct{}, len(values))
	for _, value := range values {
		role := AdminRole(strings.TrimSpace(value))
		if role == "" {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		roles = append(roles, role)
	}
	return roles
}

type actorContextKey struct{}

func ContextWithActor(ctx context.Context, actor AdminActor) context.Context {
	return context.WithValue(ctx, actorContextKey{}, actor)
}

func ActorFromContext(ctx context.Context) (AdminActor, bool) {
	actor, ok := ctx.Value(actorContextKey{}).(AdminActor)
	return actor, ok
}
