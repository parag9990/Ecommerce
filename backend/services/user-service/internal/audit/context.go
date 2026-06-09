package audit

import "context"

type contextKey struct{}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, contextKey{}, actor)
}

func ActorFromContext(ctx context.Context) (Actor, error) {
	actor, ok := ctx.Value(contextKey{}).(Actor)
	if !ok {
		return Actor{}, ErrMissingActor
	}
	return NewActor(actor.ID, actor.Type)
}
