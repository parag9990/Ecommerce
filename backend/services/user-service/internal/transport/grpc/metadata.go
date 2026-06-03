package grpc

import (
	"context"
	"strings"

	"github.com/parag/ecommerce/backend/services/user-service/internal/audit"
	"github.com/parag/ecommerce/backend/services/user-service/internal/usecase"
	"google.golang.org/grpc/metadata"
)

const (
	metadataUserID      = "x-user-id"
	metadataServiceName = "x-service-name"
	metadataActorID     = "x-actor-id"
	metadataActorType   = "x-actor-type"
	metadataRoles       = "x-roles"
	metadataRequestID   = "x-request-id"
)

func callerFromContext(ctx context.Context) usecase.Caller {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return usecase.Caller{}
	}

	caller := usecase.Caller{
		UserID:      firstMetadataValue(md, metadataUserID),
		ServiceName: firstMetadataValue(md, metadataServiceName),
		ActorID:     firstMetadataValue(md, metadataActorID),
		ActorType:   firstMetadataValue(md, metadataActorType),
		Roles:       splitMetadataValues(md.Get(metadataRoles)),
	}
	if caller.ActorID == "" {
		if actor, err := audit.ActorFromContext(ctx); err == nil {
			caller.ActorID = actor.ID
			caller.ActorType = string(actor.Type)
		}
	}
	return caller
}

func requestIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	return firstMetadataValue(md, metadataRequestID)
}

func firstMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func actorFromMetadata(ctx context.Context) (audit.Actor, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return audit.Actor{}, false
	}

	actorID := firstMetadataValue(md, metadataActorID)
	actorType := audit.ActorType(firstMetadataValue(md, metadataActorType))
	if actorID == "" {
		actorID = firstMetadataValue(md, metadataUserID)
		actorType = audit.ActorTypeUser
	}
	if actorID == "" {
		actorID = firstMetadataValue(md, metadataServiceName)
		actorType = audit.ActorTypeService
	}

	actor, err := audit.NewActor(actorID, actorType)
	if err != nil {
		return audit.Actor{}, false
	}
	return actor, true
}

func splitMetadataValues(values []string) []string {
	roles := make([]string, 0)
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			role := strings.TrimSpace(part)
			if role != "" {
				roles = append(roles, role)
			}
		}
	}
	return roles
}
