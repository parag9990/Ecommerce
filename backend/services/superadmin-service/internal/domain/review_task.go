package domain

type ReviewTaskStatus string

const (
	ReviewTaskStatusOpen      ReviewTaskStatus = "open"
	ReviewTaskStatusApproved  ReviewTaskStatus = "approved"
	ReviewTaskStatusRejected  ReviewTaskStatus = "rejected"
	ReviewTaskStatusCancelled ReviewTaskStatus = "cancelled"
)

type CloseReviewTaskRequest struct {
	TaskType     string
	ResourceType string
	ResourceID   string
	Status       ReviewTaskStatus
	ReviewedBy   string
	Reason       string
}

func SellerKYCReviewStatusForSellerStatus(status SellerStatus) (ReviewTaskStatus, bool) {
	switch status {
	case SellerStatusActive:
		return ReviewTaskStatusApproved, true
	case SellerStatusRejected:
		return ReviewTaskStatusRejected, true
	default:
		return "", false
	}
}
