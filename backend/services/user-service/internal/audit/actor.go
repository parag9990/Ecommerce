package audit

import (
	"errors"
	"strings"
)

type ActorType string

const (
	ActorTypeUser    ActorType = "user"
	ActorTypeAdmin   ActorType = "admin"
	ActorTypeService ActorType = "service"
	ActorTypeSystem  ActorType = "system"
)

var (
	ErrMissingActor = errors.New("audit actor is required")
	ErrInvalidActor = errors.New("audit actor is invalid")
)

type Actor struct {
	ID   string
	Type ActorType
}

func NewActor(id string, actorType ActorType) (Actor, error) {
	actor := Actor{ID: strings.TrimSpace(id), Type: ActorType(strings.TrimSpace(string(actorType)))}
	if actor.ID == "" {
		return Actor{}, ErrMissingActor
	}
	if !actor.Type.Valid() {
		return Actor{}, ErrInvalidActor
	}
	return actor, nil
}

func (t ActorType) Valid() bool {
	switch t {
	case ActorTypeUser, ActorTypeAdmin, ActorTypeService, ActorTypeSystem:
		return true
	default:
		return false
	}
}

func (a Actor) AuditID() string {
	id := strings.TrimSpace(a.ID)
	switch a.Type {
	case ActorTypeService, ActorTypeSystem:
		prefix := string(a.Type) + ":"
		if strings.HasPrefix(id, prefix) {
			return id
		}
		return prefix + id
	default:
		return id
	}
}
