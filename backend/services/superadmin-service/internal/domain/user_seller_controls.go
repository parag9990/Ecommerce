package domain

import (
	"fmt"
	"strings"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
	MaxSearchLength = 128
	MinReasonLength = 10
	MaxReasonLength = 512
)

type UserStatus string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusBlocked UserStatus = "blocked"
	UserStatusDeleted UserStatus = "deleted"
)

func ParseUserStatus(value string) (UserStatus, error) {
	status := UserStatus(strings.TrimSpace(value))
	if !status.Valid() {
		return "", NewInvalidStatus("user", value)
	}
	return status, nil
}

func (s UserStatus) Valid() bool {
	switch s {
	case UserStatusActive, UserStatusBlocked, UserStatusDeleted:
		return true
	default:
		return false
	}
}

func CanUpdateUserStatus(current UserStatus, next UserStatus) bool {
	switch current {
	case UserStatusActive:
		return next == UserStatusBlocked
	case UserStatusBlocked:
		return next == UserStatusActive
	default:
		return false
	}
}

type SellerStatus string

const (
	SellerStatusDraft         SellerStatus = "draft"
	SellerStatusPendingReview SellerStatus = "pending_review"
	SellerStatusActive        SellerStatus = "active"
	SellerStatusSuspended     SellerStatus = "suspended"
	SellerStatusRejected      SellerStatus = "rejected"
)

func ParseSellerStatus(value string) (SellerStatus, error) {
	status := SellerStatus(strings.TrimSpace(value))
	if !status.Valid() {
		return "", NewInvalidStatus("seller", value)
	}
	return status, nil
}

func (s SellerStatus) Valid() bool {
	switch s {
	case SellerStatusDraft, SellerStatusPendingReview, SellerStatusActive, SellerStatusSuspended, SellerStatusRejected:
		return true
	default:
		return false
	}
}

func CanUpdateSellerStatus(current SellerStatus, next SellerStatus) bool {
	switch current {
	case SellerStatusPendingReview:
		return next == SellerStatusActive || next == SellerStatusRejected
	case SellerStatusActive:
		return next == SellerStatusSuspended
	case SellerStatusSuspended:
		return next == SellerStatusActive
	default:
		return false
	}
}

type KYCStatus string

const (
	KYCStatusPending  KYCStatus = "pending"
	KYCStatusApproved KYCStatus = "approved"
	KYCStatusRejected KYCStatus = "rejected"
)

func (s KYCStatus) Valid() bool {
	switch s {
	case KYCStatusPending, KYCStatusApproved, KYCStatusRejected:
		return true
	default:
		return false
	}
}

type Pagination struct {
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	Cursor   string `json:"cursor,omitempty"`
}

func NormalizePagination(page int, pageSize int, cursor string) Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return Pagination{Page: page, PageSize: pageSize, Cursor: strings.TrimSpace(cursor)}
}

type AdminUserListRequest struct {
	Query  string     `json:"q,omitempty"`
	Status UserStatus `json:"status,omitempty"`
	Pagination
}

func (r AdminUserListRequest) Normalize() (AdminUserListRequest, error) {
	r.Query = strings.TrimSpace(r.Query)
	if len(r.Query) > MaxSearchLength {
		return AdminUserListRequest{}, NewValidationError(fmt.Sprintf("q must be at most %d characters", MaxSearchLength))
	}
	if strings.TrimSpace(string(r.Status)) != "" && !r.Status.Valid() {
		return AdminUserListRequest{}, NewInvalidStatus("user", string(r.Status))
	}
	r.Pagination = NormalizePagination(r.Pagination.Page, r.Pagination.PageSize, r.Pagination.Cursor)
	return r, nil
}

type AdminSellerListRequest struct {
	Query  string       `json:"q,omitempty"`
	Status SellerStatus `json:"status,omitempty"`
	Pagination
}

func (r AdminSellerListRequest) Normalize() (AdminSellerListRequest, error) {
	r.Query = strings.TrimSpace(r.Query)
	if len(r.Query) > MaxSearchLength {
		return AdminSellerListRequest{}, NewValidationError(fmt.Sprintf("q must be at most %d characters", MaxSearchLength))
	}
	if strings.TrimSpace(string(r.Status)) != "" && !r.Status.Valid() {
		return AdminSellerListRequest{}, NewInvalidStatus("seller", string(r.Status))
	}
	r.Pagination = NormalizePagination(r.Pagination.Page, r.Pagination.PageSize, r.Pagination.Cursor)
	return r, nil
}

type UserProfile struct {
	UserID   string     `json:"user_id"`
	Email    string     `json:"email,omitempty"`
	Phone    string     `json:"phone,omitempty"`
	FullName string     `json:"full_name,omitempty"`
	Status   UserStatus `json:"status"`
	Roles    []string   `json:"roles,omitempty"`
}

type AdminUserListResponse struct {
	Users    []UserProfile `json:"users"`
	Page     int           `json:"page,omitempty"`
	PageSize int           `json:"page_size,omitempty"`
	Total    int64         `json:"total,omitempty"`
	Cursor   string        `json:"cursor,omitempty"`
}

type KYCDocument struct {
	DocumentID      string    `json:"document_id"`
	SellerID        string    `json:"seller_id,omitempty"`
	DocumentType    string    `json:"document_type"`
	StorageURL      string    `json:"storage_url,omitempty"`
	Status          KYCStatus `json:"status"`
	RejectionReason string    `json:"rejection_reason,omitempty"`
}

type SellerProfile struct {
	SellerID     string        `json:"seller_id"`
	UserID       string        `json:"user_id"`
	StoreName    string        `json:"store_name,omitempty"`
	DisplayName  string        `json:"display_name,omitempty"`
	GSTNumber    string        `json:"gst_number,omitempty"`
	SupportEmail string        `json:"support_email,omitempty"`
	Status       SellerStatus  `json:"status"`
	KYCDocuments []KYCDocument `json:"kyc_documents,omitempty"`
}

type AdminSellerListResponse struct {
	Sellers  []SellerProfile `json:"sellers"`
	Page     int             `json:"page,omitempty"`
	PageSize int             `json:"page_size,omitempty"`
	Total    int64           `json:"total,omitempty"`
	Cursor   string          `json:"cursor,omitempty"`
}

type StatusUpdateRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

type AdminMutationContext struct {
	Actor  AdminActor
	Reason string
}

func NormalizeMutationReason(reason string) (string, error) {
	cleaned := strings.TrimSpace(reason)
	if len(cleaned) < MinReasonLength {
		return "", NewReasonRequired()
	}
	if len(cleaned) > MaxReasonLength {
		return "", NewValidationError(fmt.Sprintf("reason must be at most %d characters", MaxReasonLength))
	}
	return cleaned, nil
}
