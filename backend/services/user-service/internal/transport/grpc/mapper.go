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
		Audit:         mapAuditInfo(user.AuditFields, user.StatusAuditFields, user.SoftDeleteFields),
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
		Audit:        mapAuditInfo(seller.AuditFields, seller.StatusAuditFields, domain.SoftDeleteFields{}),
	}
}

func mapAddress(address domain.Address) *userv1.Address {
	return &userv1.Address{
		AddressId:  address.AddressID,
		UserId:     address.UserID,
		Name:       address.Name,
		Phone:      stringValue(address.Phone),
		Line1:      address.Line1,
		Line2:      stringValue(address.Line2),
		City:       address.City,
		State:      address.State,
		PostalCode: address.PostalCode,
		Country:    address.Country,
		IsDefault:  address.IsDefault,
		CreatedAt:  toProtoTime(address.CreatedAt),
		UpdatedAt:  toProtoTime(address.UpdatedAt),
		Audit:      mapAuditInfo(address.AuditFields, domain.StatusAuditFields{}, address.SoftDeleteFields),
	}
}

func mapKYCDocument(document domain.KYCDocument) *userv1.KYCDocument {
	return &userv1.KYCDocument{
		DocumentId:      document.DocumentID,
		SellerId:        document.SellerID,
		DocumentType:    string(document.DocumentType),
		StorageUrl:      document.StorageURL,
		Status:          string(document.Status),
		ReviewedBy:      stringValue(document.ReviewedBy),
		ReviewedAt:      toProtoTimePtr(document.ReviewedAt),
		RejectionReason: stringValue(document.RejectionReason),
		CreatedAt:       toProtoTime(document.CreatedAt),
		UpdatedAt:       toProtoTime(document.UpdatedAt),
		Audit:           mapAuditInfo(document.AuditFields, document.StatusAuditFields, domain.SoftDeleteFields{}),
	}
}

func mapAuditInfo(audit domain.AuditFields, status domain.StatusAuditFields, softDelete domain.SoftDeleteFields) *userv1.AuditInfo {
	return &userv1.AuditInfo{
		CreatedBy:       audit.CreatedBy,
		UpdatedBy:       audit.UpdatedBy,
		CreatedAt:       toProtoTime(audit.CreatedAt),
		UpdatedAt:       toProtoTime(audit.UpdatedAt),
		StatusChangedBy: stringValue(status.StatusChangedBy),
		StatusChangedAt: toProtoTimePtr(status.StatusChangedAt),
		StatusReason:    stringValue(status.StatusReason),
		DeletedBy:       stringValue(softDelete.DeletedBy),
		DeletedAt:       toProtoTimePtr(softDelete.DeletedAt),
	}
}

func mapAddressList(addresses []domain.Address, page int32, pageSize int32) *userv1.AddressListResponse {
	mapped := make([]*userv1.Address, 0, len(addresses))
	for _, address := range addresses {
		mapped = append(mapped, mapAddress(address))
	}
	return &userv1.AddressListResponse{
		Addresses: mapped,
		Page:      page,
		PageSize:  pageSize,
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
