package events

type OrderPayload struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Status  string `json:"status,omitempty"`
}

type PaymentPayload struct {
	OrderID       string `json:"order_id"`
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	PaymentStatus string `json:"payment_status,omitempty"`
	Amount        string `json:"amount"`
}

type UserPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}
