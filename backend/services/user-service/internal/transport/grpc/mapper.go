package grpc

import (
	"time"

	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
	userv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/user/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapUserProfile(user domain.User) *userv1.UserProfile {
	return &userv1.UserProfile{
		UserId:        user.UserID,
		AuthAccountId: user.AuthAccountID,
		Email:         user.Email,
		Phone:         stringValue(user.Phone),
		FullName:      user.FullName,
		AvatarUrl:     stringValue(user.AvatarURL),
		Status:        string(user.Status),
		CreatedAt:     toProtoTime(user.CreatedAt),
		UpdatedAt:     toProtoTime(user.UpdatedAt),
	}
}

func mapSellerProfile(seller domain.SellerProfile) *userv1.SellerProfile {
	return &userv1.SellerProfile{
		SellerId:     seller.SellerID,
		UserId:       seller.UserID,
		StoreName:    seller.StoreName,
		DisplayName:  stringValue(seller.DisplayName),
		GstNumber:    stringValue(seller.GSTNumber),
		SupportEmail: stringValue(seller.SupportEmail),
		Status:       string(seller.Status),
		ApprovedAt:   toProtoTimePtr(seller.ApprovedAt),
		CreatedAt:    toProtoTime(seller.CreatedAt),
		UpdatedAt:    toProtoTime(seller.UpdatedAt),
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func toProtoTime(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}
	return timestamppb.New(value.UTC())
}

func toProtoTimePtr(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return toProtoTime(*value)
}
