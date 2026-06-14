package retry

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	AttemptExchange = "notification.delivery.attempt"
	AttemptQueue    = "notification.delivery.attempt.v1"
	RetryExchange   = "notification.delivery.retry"
	DeadExchange    = "notification.delivery.dlx"
	DeadQueue       = "notification.delivery.dlq.v1"
)

type retryTier struct {
	Route string
	Queue string
	Delay time.Duration
}

func policyTiers(policy Policy) []retryTier {
	tiers := make([]retryTier, 0, len(policy.Delays))
	for _, delay := range policy.Delays {
		route := routeForDelay(delay)
		tiers = append(tiers, retryTier{
			Route: route,
			Queue: "notification.delivery.retry." + route[len("after."):] + ".v1",
			Delay: delay,
		})
	}
	return tiers
}

func DeclareTopology(ch *amqp.Channel, policy Policy) error {
	if ch == nil {
		return fmt.Errorf("notification retry RabbitMQ channel is required")
	}
	for _, exchange := range []string{AttemptExchange, RetryExchange, DeadExchange} {
		if err := ch.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare notification retry exchange %q: %w", exchange, err)
		}
	}
	attempt, err := ch.QueueDeclare(AttemptQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    DeadExchange,
		"x-dead-letter-routing-key": "rejected",
	})
	if err != nil {
		return fmt.Errorf("declare notification attempt queue: %w", err)
	}
	if err := ch.QueueBind(attempt.Name, "send", AttemptExchange, false, nil); err != nil {
		return fmt.Errorf("bind notification attempt queue: %w", err)
	}
	for _, tier := range policyTiers(policy) {
		queue, err := ch.QueueDeclare(tier.Queue, true, false, false, false, amqp.Table{
			"x-message-ttl":             tier.Delay.Milliseconds(),
			"x-dead-letter-exchange":    AttemptExchange,
			"x-dead-letter-routing-key": "send",
		})
		if err != nil {
			return fmt.Errorf("declare notification retry queue %q: %w", tier.Queue, err)
		}
		if err := ch.QueueBind(queue.Name, tier.Route, RetryExchange, false, nil); err != nil {
			return fmt.Errorf("bind notification retry queue %q: %w", tier.Queue, err)
		}
	}
	dead, err := ch.QueueDeclare(DeadQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare notification DLQ: %w", err)
	}
	for _, route := range []string{"failed", "invalid", "security", "rejected"} {
		if err := ch.QueueBind(dead.Name, route, DeadExchange, false, nil); err != nil {
			return fmt.Errorf("bind notification DLQ route %q: %w", route, err)
		}
	}
	return nil
}
