package domain

type OrderPaymentStatus string

const (
	OrderPaymentStatusPendingPayment    OrderPaymentStatus = "pending_payment"
	OrderPaymentStatusPaymentAuthorized OrderPaymentStatus = "payment_authorized"
	OrderPaymentStatusPaid              OrderPaymentStatus = "paid"
	OrderPaymentStatusPaymentFailed     OrderPaymentStatus = "payment_failed"
	OrderPaymentStatusPartiallyRefunded OrderPaymentStatus = "partially_refunded"
	OrderPaymentStatusRefunded          OrderPaymentStatus = "refunded"
)

type OrderStateMapping struct {
	PaymentStatus PaymentStatus
	OrderStatus   OrderPaymentStatus
	Explanation   string
}

func SuggestedOrderStatus(status PaymentStatus) (OrderPaymentStatus, bool) {
	for _, mapping := range orderStateMappings {
		if mapping.PaymentStatus == status {
			return mapping.OrderStatus, true
		}
	}
	return "", false
}

func OrderStateMappings() []OrderStateMapping {
	mappings := make([]OrderStateMapping, 0, len(orderStateMappings))
	for _, mapping := range orderStateMappings {
		mappings = append(mappings, mapping)
	}
	return mappings
}

var orderStateMappings = []OrderStateMapping{
	{
		PaymentStatus: PaymentStatusInitiated,
		OrderStatus:   OrderPaymentStatusPendingPayment,
		Explanation:   "Order exists and payment has started.",
	},
	{
		PaymentStatus: PaymentStatusRequiresAction,
		OrderStatus:   OrderPaymentStatusPendingPayment,
		Explanation:   "User action is still pending.",
	},
	{
		PaymentStatus: PaymentStatusAuthorized,
		OrderStatus:   OrderPaymentStatusPaymentAuthorized,
		Explanation:   "Provider has authorized amount, but paid should wait for capture.",
	},
	{
		PaymentStatus: PaymentStatusCaptured,
		OrderStatus:   OrderPaymentStatusPaid,
		Explanation:   "Payment is successful and fulfillment may start.",
	},
	{
		PaymentStatus: PaymentStatusFailed,
		OrderStatus:   OrderPaymentStatusPaymentFailed,
		Explanation:   "Payment attempt failed; inventory release or retry flow may run.",
	},
	{
		PaymentStatus: PaymentStatusRetryAllowed,
		OrderStatus:   OrderPaymentStatusPaymentFailed,
		Explanation:   "Retry option may be shown, but order is still not paid.",
	},
	{
		PaymentStatus: PaymentStatusPartiallyRefunded,
		OrderStatus:   OrderPaymentStatusPartiallyRefunded,
		Explanation:   "Order was paid and a partial refund has completed.",
	},
	{
		PaymentStatus: PaymentStatusRefunded,
		OrderStatus:   OrderPaymentStatusRefunded,
		Explanation:   "Full amount has been refunded.",
	},
}
