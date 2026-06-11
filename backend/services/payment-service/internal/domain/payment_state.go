package domain

import (
	"sort"
	"strings"
)

type PaymentStatus string

const (
	PaymentStatusUnspecified       PaymentStatus = ""
	PaymentStatusInitiated         PaymentStatus = "initiated"
	PaymentStatusRequiresAction    PaymentStatus = "requires_action"
	PaymentStatusAuthorized        PaymentStatus = "authorized"
	PaymentStatusCaptured          PaymentStatus = "captured"
	PaymentStatusFailed            PaymentStatus = "failed"
	PaymentStatusRetryAllowed      PaymentStatus = "retry_allowed"
	PaymentStatusPartiallyRefunded PaymentStatus = "partially_refunded"
	PaymentStatusRefunded          PaymentStatus = "refunded"
)

type PaymentEvent string

const (
	PaymentEventAttemptCreated         PaymentEvent = "payment_attempt_created"
	PaymentEventProviderRequiresAction PaymentEvent = "provider_requires_action"
	PaymentEventProviderAuthorized     PaymentEvent = "provider_authorized"
	PaymentEventProviderCaptured       PaymentEvent = "provider_captured"
	PaymentEventProviderFailed         PaymentEvent = "provider_failed"
	PaymentEventActionCompleted        PaymentEvent = "action_completed"
	PaymentEventActionFailed           PaymentEvent = "action_failed"
	PaymentEventAuthorizationFailed    PaymentEvent = "authorization_failed"
	PaymentEventPartialRefundSucceeded PaymentEvent = "partial_refund_succeeded"
	PaymentEventFullRefundSucceeded    PaymentEvent = "full_refund_succeeded"
	PaymentEventRetryRequested         PaymentEvent = "retry_requested"
	PaymentEventNewAttemptCreated      PaymentEvent = "new_attempt_created"
)

type PaymentStatusDefinition struct {
	Status          PaymentStatus
	Meaning         string
	UserFacing      string
	FinalForAttempt bool
	Helper          bool
	Persisted       bool
}

type TransitionRule struct {
	From          PaymentStatus
	Event         PaymentEvent
	To            PaymentStatus
	TriggerSource string
	Notes         string
}

type InvalidTransition struct {
	From   PaymentStatus
	To     PaymentStatus
	Reason string
}

type TransitionDecision struct {
	From    PaymentStatus
	To      PaymentStatus
	Event   PaymentEvent
	Allowed bool
	Noop    bool
	Reason  string
}

type StateMachineDefinition struct {
	Statuses         []PaymentStatusDefinition
	Transitions      []TransitionRule
	InvalidExamples  []InvalidTransition
	ProviderMappings []ProviderEventMapping
	OrderMappings    []OrderStateMapping
}

type StateMachine struct {
	statuses          map[PaymentStatus]PaymentStatusDefinition
	transitions       []TransitionRule
	transitionsByFrom map[PaymentStatus][]TransitionRule
	events            map[PaymentEvent]PaymentStatus
	invalidExamples   []InvalidTransition
}

func NewStateMachine() *StateMachine {
	statuses := make(map[PaymentStatus]PaymentStatusDefinition, len(paymentStatusDefinitions))
	for _, definition := range paymentStatusDefinitions {
		statuses[definition.Status] = definition
	}

	transitions := append([]TransitionRule(nil), defaultTransitionRules...)
	transitionsByFrom := make(map[PaymentStatus][]TransitionRule, len(transitions))
	for _, rule := range transitions {
		transitionsByFrom[rule.From] = append(transitionsByFrom[rule.From], rule)
	}

	events := make(map[PaymentEvent]PaymentStatus, len(paymentEventTargets))
	for event, status := range paymentEventTargets {
		events[event] = status
	}

	return &StateMachine{
		statuses:          statuses,
		transitions:       transitions,
		transitionsByFrom: transitionsByFrom,
		events:            events,
		invalidExamples:   append([]InvalidTransition(nil), documentedInvalidTransitions...),
	}
}

func (m *StateMachine) Definition() StateMachineDefinition {
	if m == nil {
		m = NewStateMachine()
	}

	statuses := make([]PaymentStatusDefinition, 0, len(paymentStatusDefinitions))
	for _, definition := range paymentStatusDefinitions {
		statuses = append(statuses, definition)
	}

	return StateMachineDefinition{
		Statuses:         statuses,
		Transitions:      append([]TransitionRule(nil), m.transitions...),
		InvalidExamples:  append([]InvalidTransition(nil), m.invalidExamples...),
		ProviderMappings: ProviderEventMappings(),
		OrderMappings:    OrderStateMappings(),
	}
}

func (m *StateMachine) IsKnownStatus(status PaymentStatus) bool {
	if m == nil {
		m = NewStateMachine()
	}
	if status == PaymentStatusUnspecified {
		return true
	}
	_, ok := m.statuses[status]
	return ok
}

func (m *StateMachine) IsKnownEvent(event PaymentEvent) bool {
	if m == nil {
		m = NewStateMachine()
	}
	_, ok := m.events[event]
	return ok
}

func (m *StateMachine) CanTransition(from PaymentStatus, to PaymentStatus) bool {
	if m == nil {
		m = NewStateMachine()
	}
	if !m.IsKnownStatus(from) || !m.IsKnownStatus(to) || to == PaymentStatusUnspecified {
		return false
	}
	if from == to {
		return true
	}
	for _, rule := range m.transitionsByFrom[from] {
		if rule.To == to {
			return true
		}
	}
	return false
}

func (m *StateMachine) ValidateTransition(from PaymentStatus, to PaymentStatus) error {
	if m == nil {
		m = NewStateMachine()
	}
	if !m.IsKnownStatus(from) {
		return InvalidPaymentStatusError{Status: from}
	}
	if !m.IsKnownStatus(to) || to == PaymentStatusUnspecified {
		return InvalidPaymentStatusError{Status: to}
	}
	if !m.CanTransition(from, to) {
		return InvalidPaymentTransitionError{From: from, To: to}
	}
	return nil
}

func (m *StateMachine) NextStatusForEvent(from PaymentStatus, event PaymentEvent) (PaymentStatus, bool, error) {
	if m == nil {
		m = NewStateMachine()
	}
	if !m.IsKnownStatus(from) {
		return PaymentStatusUnspecified, false, InvalidPaymentStatusError{Status: from}
	}
	to, ok := m.events[event]
	if !ok {
		return PaymentStatusUnspecified, false, InvalidPaymentEventError{Event: event}
	}
	if !m.CanTransition(from, to) {
		return to, false, InvalidPaymentTransitionError{From: from, To: to, Event: event}
	}
	return to, from == to, nil
}

func (m *StateMachine) EvaluateTransition(from PaymentStatus, to PaymentStatus) TransitionDecision {
	if err := m.ValidateTransition(from, to); err != nil {
		return TransitionDecision{
			From:    from,
			To:      to,
			Allowed: false,
			Reason:  err.Error(),
		}
	}
	reason := "transition is allowed"
	if from == to {
		reason = "duplicate state update is an idempotent no-op"
	}
	return TransitionDecision{
		From:    from,
		To:      to,
		Allowed: true,
		Noop:    from == to,
		Reason:  reason,
	}
}

func (m *StateMachine) EvaluateEvent(from PaymentStatus, event PaymentEvent) TransitionDecision {
	to, noop, err := m.NextStatusForEvent(from, event)
	if err != nil {
		return TransitionDecision{
			From:    from,
			To:      to,
			Event:   event,
			Allowed: false,
			Reason:  err.Error(),
		}
	}
	reason := "event transition is allowed"
	if noop {
		reason = "duplicate event maps to current state and is an idempotent no-op"
	}
	return TransitionDecision{
		From:    from,
		To:      to,
		Event:   event,
		Allowed: true,
		Noop:    noop,
		Reason:  reason,
	}
}

func (m *StateMachine) AllowedNextStatuses(from PaymentStatus) ([]PaymentStatus, error) {
	if m == nil {
		m = NewStateMachine()
	}
	if !m.IsKnownStatus(from) {
		return nil, InvalidPaymentStatusError{Status: from}
	}
	next := make([]PaymentStatus, 0, len(m.transitionsByFrom[from])+1)
	if from != PaymentStatusUnspecified {
		next = append(next, from)
	}
	for _, rule := range m.transitionsByFrom[from] {
		next = append(next, rule.To)
	}
	sort.Slice(next, func(i, j int) bool {
		return paymentStatusRank(next[i]) < paymentStatusRank(next[j])
	})
	return next, nil
}

func ParsePaymentStatus(raw string) (PaymentStatus, error) {
	status := PaymentStatus(strings.ToLower(strings.TrimSpace(raw)))
	if NewStateMachine().IsKnownStatus(status) {
		return status, nil
	}
	return status, InvalidPaymentStatusError{Status: status}
}

func ParsePaymentEvent(raw string) (PaymentEvent, error) {
	event := PaymentEvent(strings.ToLower(strings.TrimSpace(raw)))
	if NewStateMachine().IsKnownEvent(event) {
		return event, nil
	}
	return event, InvalidPaymentEventError{Event: event}
}

func IsPersistedPaymentStatus(status PaymentStatus) bool {
	definition, ok := NewStateMachine().statuses[status]
	return ok && definition.Persisted
}

func CanTransitionPayment(from PaymentStatus, to PaymentStatus) bool {
	return NewStateMachine().CanTransition(from, to)
}

func ValidatePaymentTransition(from PaymentStatus, to PaymentStatus) error {
	return NewStateMachine().ValidateTransition(from, to)
}

func NextStatusFromPaymentEvent(from PaymentStatus, event PaymentEvent) (PaymentStatus, bool, error) {
	return NewStateMachine().NextStatusForEvent(from, event)
}

func paymentStatusRank(status PaymentStatus) int {
	for i, known := range paymentStatusOrder {
		if known == status {
			return i
		}
	}
	return len(paymentStatusOrder)
}

var paymentStatusOrder = []PaymentStatus{
	PaymentStatusUnspecified,
	PaymentStatusInitiated,
	PaymentStatusRequiresAction,
	PaymentStatusAuthorized,
	PaymentStatusCaptured,
	PaymentStatusFailed,
	PaymentStatusRetryAllowed,
	PaymentStatusPartiallyRefunded,
	PaymentStatusRefunded,
}

var paymentStatusDefinitions = []PaymentStatusDefinition{
	{
		Status:          PaymentStatusInitiated,
		Meaning:         "Payment process has started and a provider intent or attempt is being created.",
		UserFacing:      "Payment started.",
		FinalForAttempt: false,
		Persisted:       true,
	},
	{
		Status:          PaymentStatusRequiresAction,
		Meaning:         "Provider requires user action such as OTP, 3DS, redirect, or UPI approval.",
		UserFacing:      "User action required to complete payment.",
		FinalForAttempt: false,
		Persisted:       true,
	},
	{
		Status:          PaymentStatusAuthorized,
		Meaning:         "Provider authorized the amount, but capture is not complete.",
		UserFacing:      "Payment approved and awaiting confirmation.",
		FinalForAttempt: false,
		Persisted:       true,
	},
	{
		Status:          PaymentStatusCaptured,
		Meaning:         "Money has been captured or charged by the provider.",
		UserFacing:      "Payment successful.",
		FinalForAttempt: true,
		Persisted:       true,
	},
	{
		Status:          PaymentStatusFailed,
		Meaning:         "Provider declined, expired, cancelled, or otherwise failed this attempt.",
		UserFacing:      "Payment failed.",
		FinalForAttempt: true,
		Persisted:       true,
	},
	{
		Status:          PaymentStatusRetryAllowed,
		Meaning:         "Retry policy permits a new attempt for the same order after failure.",
		UserFacing:      "Payment can be retried.",
		FinalForAttempt: false,
		Helper:          true,
		Persisted:       false,
	},
	{
		Status:          PaymentStatusPartiallyRefunded,
		Meaning:         "Captured payment has been partially refunded.",
		UserFacing:      "Partial refund completed.",
		FinalForAttempt: false,
		Persisted:       true,
	},
	{
		Status:          PaymentStatusRefunded,
		Meaning:         "Captured payment has been fully refunded.",
		UserFacing:      "Full refund completed.",
		FinalForAttempt: true,
		Persisted:       true,
	},
}

var defaultTransitionRules = []TransitionRule{
	{
		From:          PaymentStatusUnspecified,
		Event:         PaymentEventAttemptCreated,
		To:            PaymentStatusInitiated,
		TriggerSource: "Order Service / Payment Service",
		Notes:         "Checkout creates the first payment attempt.",
	},
	{
		From:          PaymentStatusInitiated,
		Event:         PaymentEventProviderRequiresAction,
		To:            PaymentStatusRequiresAction,
		TriggerSource: "Provider response/webhook",
		Notes:         "OTP, 3DS, redirect, or UPI approval required.",
	},
	{
		From:          PaymentStatusInitiated,
		Event:         PaymentEventProviderAuthorized,
		To:            PaymentStatusAuthorized,
		TriggerSource: "Provider response/webhook",
		Notes:         "Amount authorized, capture pending.",
	},
	{
		From:          PaymentStatusInitiated,
		Event:         PaymentEventProviderCaptured,
		To:            PaymentStatusCaptured,
		TriggerSource: "Provider webhook",
		Notes:         "Some providers authorize and capture in one step.",
	},
	{
		From:          PaymentStatusInitiated,
		Event:         PaymentEventProviderFailed,
		To:            PaymentStatusFailed,
		TriggerSource: "Provider response/webhook",
		Notes:         "Declined, expired, or cancelled.",
	},
	{
		From:          PaymentStatusRequiresAction,
		Event:         PaymentEventActionCompleted,
		To:            PaymentStatusAuthorized,
		TriggerSource: "Provider webhook",
		Notes:         "User action completed successfully.",
	},
	{
		From:          PaymentStatusRequiresAction,
		Event:         PaymentEventProviderAuthorized,
		To:            PaymentStatusAuthorized,
		TriggerSource: "Provider webhook",
		Notes:         "Provider reports authorization after required user action.",
	},
	{
		From:          PaymentStatusRequiresAction,
		Event:         PaymentEventActionFailed,
		To:            PaymentStatusFailed,
		TriggerSource: "Provider webhook",
		Notes:         "User cancelled or action expired.",
	},
	{
		From:          PaymentStatusRequiresAction,
		Event:         PaymentEventProviderFailed,
		To:            PaymentStatusFailed,
		TriggerSource: "Provider webhook",
		Notes:         "Provider reports failed payment after required user action.",
	},
	{
		From:          PaymentStatusRequiresAction,
		Event:         PaymentEventProviderCaptured,
		To:            PaymentStatusCaptured,
		TriggerSource: "Provider webhook",
		Notes:         "Provider captures payment immediately after required user action.",
	},
	{
		From:          PaymentStatusAuthorized,
		Event:         PaymentEventProviderCaptured,
		To:            PaymentStatusCaptured,
		TriggerSource: "Provider webhook",
		Notes:         "Money captured.",
	},
	{
		From:          PaymentStatusAuthorized,
		Event:         PaymentEventAuthorizationFailed,
		To:            PaymentStatusFailed,
		TriggerSource: "Provider webhook",
		Notes:         "Authorization voided or expired.",
	},
	{
		From:          PaymentStatusAuthorized,
		Event:         PaymentEventProviderFailed,
		To:            PaymentStatusFailed,
		TriggerSource: "Provider webhook",
		Notes:         "Provider reports a failed authorized payment before capture.",
	},
	{
		From:          PaymentStatusCaptured,
		Event:         PaymentEventPartialRefundSucceeded,
		To:            PaymentStatusPartiallyRefunded,
		TriggerSource: "Provider refund webhook",
		Notes:         "Refund amount is less than captured amount.",
	},
	{
		From:          PaymentStatusCaptured,
		Event:         PaymentEventFullRefundSucceeded,
		To:            PaymentStatusRefunded,
		TriggerSource: "Provider refund webhook",
		Notes:         "Refund amount equals captured amount.",
	},
	{
		From:          PaymentStatusPartiallyRefunded,
		Event:         PaymentEventPartialRefundSucceeded,
		To:            PaymentStatusPartiallyRefunded,
		TriggerSource: "Provider refund webhook",
		Notes:         "Another partial refund succeeded and full amount is not yet refunded.",
	},
	{
		From:          PaymentStatusPartiallyRefunded,
		Event:         PaymentEventFullRefundSucceeded,
		To:            PaymentStatusRefunded,
		TriggerSource: "Provider refund webhook",
		Notes:         "Remaining amount refunded.",
	},
	{
		From:          PaymentStatusFailed,
		Event:         PaymentEventRetryRequested,
		To:            PaymentStatusRetryAllowed,
		TriggerSource: "User/Order Service policy",
		Notes:         "Retry policy permits a new attempt.",
	},
	{
		From:          PaymentStatusRetryAllowed,
		Event:         PaymentEventNewAttemptCreated,
		To:            PaymentStatusInitiated,
		TriggerSource: "Payment Service",
		Notes:         "New payment attempt starts.",
	},
}

var paymentEventTargets = map[PaymentEvent]PaymentStatus{
	PaymentEventAttemptCreated:         PaymentStatusInitiated,
	PaymentEventProviderRequiresAction: PaymentStatusRequiresAction,
	PaymentEventProviderAuthorized:     PaymentStatusAuthorized,
	PaymentEventProviderCaptured:       PaymentStatusCaptured,
	PaymentEventProviderFailed:         PaymentStatusFailed,
	PaymentEventActionCompleted:        PaymentStatusAuthorized,
	PaymentEventActionFailed:           PaymentStatusFailed,
	PaymentEventAuthorizationFailed:    PaymentStatusFailed,
	PaymentEventPartialRefundSucceeded: PaymentStatusPartiallyRefunded,
	PaymentEventFullRefundSucceeded:    PaymentStatusRefunded,
	PaymentEventRetryRequested:         PaymentStatusRetryAllowed,
	PaymentEventNewAttemptCreated:      PaymentStatusInitiated,
}

var documentedInvalidTransitions = []InvalidTransition{
	{
		From:   PaymentStatusFailed,
		To:     PaymentStatusCaptured,
		Reason: "A failed attempt must not be converted to success; retry creates a new attempt.",
	},
	{
		From:   PaymentStatusCaptured,
		To:     PaymentStatusAuthorized,
		Reason: "Captured payment is already successful and cannot move backward.",
	},
	{
		From:   PaymentStatusRefunded,
		To:     PaymentStatusCaptured,
		Reason: "A fully refunded payment must not become paid again.",
	},
	{
		From:   PaymentStatusInitiated,
		To:     PaymentStatusRefunded,
		Reason: "Refunds are valid only after capture.",
	},
	{
		From:   PaymentStatusAuthorized,
		To:     PaymentStatusPartiallyRefunded,
		Reason: "Refunds are valid only after capture.",
	},
	{
		From:   PaymentStatusPartiallyRefunded,
		To:     PaymentStatusCaptured,
		Reason: "Partial refund cannot be rolled back to captured-only state.",
	},
	{
		From:   PaymentStatusRefunded,
		To:     PaymentStatusPartiallyRefunded,
		Reason: "Full refund is final.",
	},
}
