package client

import (
	"context"
	"fmt"
	"strings"
)

type ModerationDecision string

const (
	ModerationAutoPublish    ModerationDecision = "auto_publish"
	ModerationReviewRequired ModerationDecision = "review_required"
)

type CMSClient interface {
	CanSellerManageCatalog(ctx context.Context, sellerID string) (bool, error)
	GetProductModerationDecision(ctx context.Context, sellerID string, categoryID string) (ModerationDecision, error)
}

type StaticCMSPolicy struct {
	CatalogManagementAllowed bool
	ModerationDecision       ModerationDecision
}

type StaticCMSPolicyClient struct {
	policy StaticCMSPolicy
}

func NewStaticCMSPolicyClient(policy StaticCMSPolicy) (*StaticCMSPolicyClient, error) {
	policy.ModerationDecision = NormalizeModerationDecision(policy.ModerationDecision)
	if !policy.ModerationDecision.Valid() {
		return nil, fmt.Errorf("cms moderation decision must be one of %s or %s", ModerationAutoPublish, ModerationReviewRequired)
	}
	return &StaticCMSPolicyClient{policy: policy}, nil
}

func (c *StaticCMSPolicyClient) CanSellerManageCatalog(ctx context.Context, sellerID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return c.policy.CatalogManagementAllowed, nil
}

func (c *StaticCMSPolicyClient) GetProductModerationDecision(ctx context.Context, sellerID string, categoryID string) (ModerationDecision, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return c.policy.ModerationDecision, nil
}

func NormalizeModerationDecision(decision ModerationDecision) ModerationDecision {
	return ModerationDecision(strings.ToLower(strings.TrimSpace(string(decision))))
}

func (d ModerationDecision) Valid() bool {
	switch NormalizeModerationDecision(d) {
	case ModerationAutoPublish, ModerationReviewRequired:
		return true
	default:
		return false
	}
}
