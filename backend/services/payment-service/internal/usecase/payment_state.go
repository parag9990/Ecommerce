package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
)

var (
	ErrTransitionTargetRequired = errors.New("transition target or event is required")
	ErrTransitionTargetConflict = errors.New("transition target does not match event target")
)

type PaymentStateMachine interface {
	Definition() domain.StateMachineDefinition
	EvaluateTransition(from domain.PaymentStatus, to domain.PaymentStatus) domain.TransitionDecision
	EvaluateEvent(from domain.PaymentStatus, event domain.PaymentEvent) domain.TransitionDecision
	AllowedNextStatuses(from domain.PaymentStatus) ([]domain.PaymentStatus, error)
}

type PaymentStateUsecase struct {
	machine PaymentStateMachine
	logger  *slog.Logger
}

type EvaluateTransitionInput struct {
	From          string
	To            string
	Event         string
	ProviderEvent string
}

type EvaluateTransitionOutput struct {
	Allowed              bool
	Noop                 bool
	From                 domain.PaymentStatus
	To                   domain.PaymentStatus
	Event                domain.PaymentEvent
	ProviderEvent        string
	SuggestedOrderStatus domain.OrderPaymentStatus
	Reason               string
}

type ProviderEventOutput struct {
	ProviderEvent string
	InternalEvent domain.PaymentEvent
	NextStatus    domain.PaymentStatus
}

func NewPaymentStateUsecase(machine PaymentStateMachine, logger *slog.Logger) (*PaymentStateUsecase, error) {
	if machine == nil {
		machine = domain.NewStateMachine()
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &PaymentStateUsecase{
		machine: machine,
		logger:  logger,
	}, nil
}

func (u *PaymentStateUsecase) Definition(ctx context.Context) (domain.StateMachineDefinition, error) {
	if err := ctx.Err(); err != nil {
		return domain.StateMachineDefinition{}, err
	}
	return u.machine.Definition(), nil
}

func (u *PaymentStateUsecase) EvaluateTransition(ctx context.Context, input EvaluateTransitionInput) (EvaluateTransitionOutput, error) {
	if err := ctx.Err(); err != nil {
		return EvaluateTransitionOutput{}, err
	}

	from, err := domain.ParsePaymentStatus(input.From)
	if err != nil {
		u.logValidationFailure(ctx, input, err)
		return EvaluateTransitionOutput{}, err
	}

	event, providerEvent, err := normalizeEvent(input.Event, input.ProviderEvent)
	if err != nil {
		u.logValidationFailure(ctx, input, err)
		return EvaluateTransitionOutput{}, err
	}

	if event == "" && strings.TrimSpace(input.To) == "" {
		err := ErrTransitionTargetRequired
		u.logValidationFailure(ctx, input, err)
		return EvaluateTransitionOutput{}, err
	}

	var decision domain.TransitionDecision
	if event != "" {
		decision = u.machine.EvaluateEvent(from, event)
		if !decision.Allowed {
			u.logValidationFailure(ctx, input, domain.InvalidPaymentTransitionError{
				From:  from,
				To:    decision.To,
				Event: event,
			})
			return transitionOutput(decision, providerEvent), domain.InvalidPaymentTransitionError{
				From:  from,
				To:    decision.To,
				Event: event,
			}
		}
		if strings.TrimSpace(input.To) != "" {
			to, err := domain.ParsePaymentStatus(input.To)
			if err != nil {
				u.logValidationFailure(ctx, input, err)
				return EvaluateTransitionOutput{}, err
			}
			if to != decision.To {
				err := fmt.Errorf("%w: event %s maps to %s, got %s", ErrTransitionTargetConflict, event, decision.To, to)
				u.logValidationFailure(ctx, input, err)
				return EvaluateTransitionOutput{}, err
			}
		}
		return transitionOutput(decision, providerEvent), nil
	}

	to, err := domain.ParsePaymentStatus(input.To)
	if err != nil {
		u.logValidationFailure(ctx, input, err)
		return EvaluateTransitionOutput{}, err
	}
	decision = u.machine.EvaluateTransition(from, to)
	if !decision.Allowed {
		u.logValidationFailure(ctx, input, domain.InvalidPaymentTransitionError{From: from, To: to})
		return transitionOutput(decision, providerEvent), domain.InvalidPaymentTransitionError{From: from, To: to}
	}
	return transitionOutput(decision, providerEvent), nil
}

func (u *PaymentStateUsecase) NormalizeProviderEvent(ctx context.Context, providerEvent string) (ProviderEventOutput, error) {
	if err := ctx.Err(); err != nil {
		return ProviderEventOutput{}, err
	}
	mapping, err := domain.NormalizeProviderEvent(providerEvent)
	if err != nil {
		return ProviderEventOutput{}, err
	}
	return ProviderEventOutput{
		ProviderEvent: mapping.ProviderEvent,
		InternalEvent: mapping.InternalEvent,
		NextStatus:    mapping.NextStatus,
	}, nil
}

func (u *PaymentStateUsecase) AllowedNextStatuses(ctx context.Context, from string) ([]domain.PaymentStatus, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	status, err := domain.ParsePaymentStatus(from)
	if err != nil {
		return nil, err
	}
	return u.machine.AllowedNextStatuses(status)
}

func normalizeEvent(event string, providerEvent string) (domain.PaymentEvent, string, error) {
	event = strings.TrimSpace(event)
	providerEvent = strings.TrimSpace(providerEvent)
	if event != "" && providerEvent != "" {
		return "", "", errors.New("provide either event or provider_event, not both")
	}
	if providerEvent != "" {
		mapping, err := domain.NormalizeProviderEvent(providerEvent)
		if err != nil {
			return "", "", err
		}
		return mapping.InternalEvent, mapping.ProviderEvent, nil
	}
	if event == "" {
		return "", "", nil
	}
	parsed, err := domain.ParsePaymentEvent(event)
	if err != nil {
		return "", "", err
	}
	return parsed, "", nil
}

func transitionOutput(decision domain.TransitionDecision, providerEvent string) EvaluateTransitionOutput {
	orderStatus, _ := domain.SuggestedOrderStatus(decision.To)
	return EvaluateTransitionOutput{
		Allowed:              decision.Allowed,
		Noop:                 decision.Noop,
		From:                 decision.From,
		To:                   decision.To,
		Event:                decision.Event,
		ProviderEvent:        providerEvent,
		SuggestedOrderStatus: orderStatus,
		Reason:               decision.Reason,
	}
}

func (u *PaymentStateUsecase) logValidationFailure(ctx context.Context, input EvaluateTransitionInput, err error) {
	u.logger.WarnContext(ctx, "payment.state_machine.validation_failed",
		slog.String("from", strings.TrimSpace(input.From)),
		slog.String("to", strings.TrimSpace(input.To)),
		slog.String("event", strings.TrimSpace(input.Event)),
		slog.String("provider_event", strings.TrimSpace(input.ProviderEvent)),
		slog.String("error", err.Error()),
	)
}
