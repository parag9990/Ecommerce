package handlers

import (
	"time"

	userv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/user/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type updateUserProfileRequest struct {
	FullName  *string `json:"full_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type addressInputRequest struct {
	Name       string  `json:"name"`
	Phone      *string `json:"phone,omitempty"`
	Line1      string  `json:"line1"`
	Line2      *string `json:"line2,omitempty"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	PostalCode string  `json:"postal_code"`
	Country    string  `json:"country"`
	IsDefault  bool    `json:"is_default"`
}

type updateSellerProfileRequest struct {
	StoreName    *string `json:"store_name,omitempty"`
	DisplayName  *string `json:"display_name,omitempty"`
	GSTNumber    *string `json:"gst_number,omitempty"`
	SupportEmail *string `json:"support_email,omitempty"`
}

type userProfileResponse struct {
	UserID        string     `json:"user_id"`
	AuthAccountID string     `json:"auth_account_id,omitempty"`
	Email         string     `json:"email"`
	Phone         string     `json:"phone,omitempty"`
	FullName      string     `json:"full_name"`
	AvatarURL     string     `json:"avatar_url,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

type addressResponse struct {
	AddressID  string     `json:"address_id"`
	UserID     string     `json:"user_id,omitempty"`
	Name       string     `json:"name"`
	Phone      string     `json:"phone,omitempty"`
	Line1      string     `json:"line1"`
	Line2      string     `json:"line2,omitempty"`
	City       string     `json:"city"`
	State      string     `json:"state"`
	PostalCode string     `json:"postal_code"`
	Country    string     `json:"country"`
	IsDefault  bool       `json:"is_default"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

type addressListResponse struct {
	Addresses []addressResponse `json:"addresses"`
	Page      int32             `json:"page"`
	PageSize  int32             `json:"page_size"`
}

type sellerProfileResponse struct {
	SellerID     string     `json:"seller_id"`
	UserID       string     `json:"user_id,omitempty"`
	StoreName    string     `json:"store_name"`
	DisplayName  string     `json:"display_name,omitempty"`
	GSTNumber    string     `json:"gst_number,omitempty"`
	SupportEmail string     `json:"support_email,omitempty"`
	Status       string     `json:"status"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type successResponse struct {
	Success bool `json:"success"`
}

func mapUserProfile(profile *userv1.UserProfile) userProfileResponse {
	if profile == nil {
		return userProfileResponse{}
	}
	return userProfileResponse{
		UserID:        profile.GetUserId(),
		AuthAccountID: profile.GetAuthAccountId(),
		Email:         profile.GetEmail(),
		Phone:         profile.GetPhone(),
		FullName:      profile.GetFullName(),
		AvatarURL:     profile.GetAvatarUrl(),
		Status:        profile.GetStatus(),
		CreatedAt:     protoTime(profile.GetCreatedAt()),
		UpdatedAt:     protoTime(profile.GetUpdatedAt()),
	}
}

func mapAddress(address *userv1.Address) addressResponse {
	if address == nil {
		return addressResponse{}
	}
	return addressResponse{
		AddressID:  address.GetAddressId(),
		UserID:     address.GetUserId(),
		Name:       address.GetName(),
		Phone:      address.GetPhone(),
		Line1:      address.GetLine1(),
		Line2:      address.GetLine2(),
		City:       address.GetCity(),
		State:      address.GetState(),
		PostalCode: address.GetPostalCode(),
		Country:    address.GetCountry(),
		IsDefault:  address.GetIsDefault(),
		CreatedAt:  protoTime(address.GetCreatedAt()),
		UpdatedAt:  protoTime(address.GetUpdatedAt()),
	}
}

func mapAddressList(response *userv1.AddressListResponse) addressListResponse {
	if response == nil {
		return addressListResponse{Addresses: []addressResponse{}}
	}
	addresses := make([]addressResponse, 0, len(response.GetAddresses()))
	for _, address := range response.GetAddresses() {
		addresses = append(addresses, mapAddress(address))
	}
	return addressListResponse{
		Addresses: addresses,
		Page:      response.GetPage(),
		PageSize:  response.GetPageSize(),
	}
}

func mapSellerProfile(profile *userv1.SellerProfile) sellerProfileResponse {
	if profile == nil {
		return sellerProfileResponse{}
	}
	return sellerProfileResponse{
		SellerID:     profile.GetSellerId(),
		UserID:       profile.GetUserId(),
		StoreName:    profile.GetStoreName(),
		DisplayName:  profile.GetDisplayName(),
		GSTNumber:    profile.GetGstNumber(),
		SupportEmail: profile.GetSupportEmail(),
		Status:       profile.GetStatus(),
		ApprovedAt:   protoTime(profile.GetApprovedAt()),
		CreatedAt:    protoTime(profile.GetCreatedAt()),
		UpdatedAt:    protoTime(profile.GetUpdatedAt()),
	}
}

func protoTime(value *timestamppb.Timestamp) *time.Time {
	if value == nil {
		return nil
	}
	t := value.AsTime().UTC()
	return &t
}
