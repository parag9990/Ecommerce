package domain

import "time"

const (
	EventUserCreated    = "UserCreated"
	EventSellerApproved = "SellerApproved"
	EventAddressUpdated = "AddressUpdated"
)

type AddressChangeType string

const (
	AddressChangeCreated        AddressChangeType = "created"
	AddressChangeUpdated        AddressChangeType = "updated"
	AddressChangeDeleted        AddressChangeType = "deleted"
	AddressChangeDefaultChanged AddressChangeType = "default_changed"
)

type UserCreatedPayload struct {
	UserID        string    `json:"user_id"`
	AuthAccountID string    `json:"auth_account_id"`
	Status        string    `json:"status"`
	EmailHash     string    `json:"email_hash,omitempty"`
	PhoneHash     string    `json:"phone_hash,omitempty"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

type SellerApprovedPayload struct {
	SellerID       string    `json:"seller_id"`
	UserID         string    `json:"user_id"`
	StoreName      string    `json:"store_name,omitempty"`
	PreviousStatus string    `json:"previous_status"`
	CurrentStatus  string    `json:"current_status"`
	ApprovedBy     string    `json:"approved_by"`
	ApprovedAt     time.Time `json:"approved_at"`
	StatusReason   string    `json:"status_reason,omitempty"`
}

type AddressUpdatedPayload struct {
	AddressID  string    `json:"address_id"`
	UserID     string    `json:"user_id"`
	ChangeType string    `json:"change_type"`
	City       string    `json:"city,omitempty"`
	State      string    `json:"state,omitempty"`
	Country    string    `json:"country,omitempty"`
	IsDefault  bool      `json:"is_default"`
	UpdatedBy  string    `json:"updated_by"`
	UpdatedAt  time.Time `json:"updated_at"`
}
