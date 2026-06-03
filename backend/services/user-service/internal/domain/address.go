package domain

import (
	"fmt"
	"time"
)

type AddressStatus string

const (
	AddressStatusActive  AddressStatus = "active"
	AddressStatusDeleted AddressStatus = "deleted"
)

type Address struct {
	AddressID  string
	UserID     string
	Name       string
	Phone      *string
	Line1      string
	Line2      *string
	City       string
	State      string
	PostalCode string
	Country    string
	IsDefault  bool
	Status     AddressStatus
	AuditFields
	SoftDeleteFields
}

type NewAddressParams struct {
	AddressID  string
	UserID     string
	Name       string
	Phone      *string
	Line1      string
	Line2      *string
	City       string
	State      string
	PostalCode string
	Country    string
	IsDefault  bool
	CreatedBy  string
	CreatedAt  time.Time
}

type AddressPatch struct {
	Name       *string
	Phone      *string
	Line1      *string
	Line2      *string
	City       *string
	State      *string
	PostalCode *string
	Country    *string
	IsDefault  *bool
	UpdatedBy  string
	UpdatedAt  time.Time
}

func (s AddressStatus) Valid() bool {
	switch s {
	case AddressStatusActive, AddressStatusDeleted:
		return true
	default:
		return false
	}
}

func NewAddress(params NewAddressParams) (Address, error) {
	createdAt := params.CreatedAt.UTC()
	address := Address{
		AddressID:  trim(params.AddressID),
		UserID:     trim(params.UserID),
		Name:       cleanText(params.Name),
		Phone:      cleanOptionalPhone(params.Phone),
		Line1:      cleanText(params.Line1),
		Line2:      cleanOptionalText(params.Line2),
		City:       cleanText(params.City),
		State:      cleanText(params.State),
		PostalCode: trim(params.PostalCode),
		Country:    cleanCountry(params.Country),
		IsDefault:  params.IsDefault,
		Status:     AddressStatusActive,
		AuditFields: AuditFields{
			CreatedBy: trim(params.CreatedBy),
			UpdatedBy: trim(params.CreatedBy),
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
	}

	if err := address.Validate(); err != nil {
		return Address{}, err
	}

	return address, nil
}

func (a Address) Validate() error {
	var v validationCollector

	validateID(&v, "address_id", a.AddressID)
	validateID(&v, "user_id", a.UserID)
	validateRequiredSafeText(&v, "name", a.Name, 2, maxNameLength)
	validateOptionalPhone(&v, "phone", a.Phone)
	validateRequiredSafeText(&v, "line1", a.Line1, 5, maxAddressLineLength)
	validateOptionalString(&v, "line2", a.Line2, maxAddressLineLength)
	validateRequiredSafeText(&v, "city", a.City, 2, maxCityStateLength)
	validateRequiredSafeText(&v, "state", a.State, 2, maxCityStateLength)
	validateRequiredSafeText(&v, "country", a.Country, 2, maxCountryLength)
	validatePostalCode(&v, "postal_code", a.PostalCode, a.Country)
	if !a.Status.Valid() {
		v.add("status", "is not supported")
	}
	validateAuditFields(&v, a.AuditFields)
	validateSoftDeleteFields(&v, a.SoftDeleteFields, a.CreatedAt)
	if a.DeletedAt != nil && a.IsDefault {
		v.add("is_default", "deleted address cannot be default")
	}
	if a.Status == AddressStatusDeleted && a.DeletedAt == nil {
		v.add("deleted_at", "is required for deleted address")
	}
	if a.Status == AddressStatusActive && a.DeletedAt != nil {
		v.add("status", "deleted_at requires deleted status")
	}

	return v.err()
}

func (a Address) BelongsTo(userID string) bool {
	return a.UserID == trim(userID)
}

func (a Address) IsDeleted() bool {
	return a.Status == AddressStatusDeleted || a.DeletedAt != nil
}

func (a *Address) ApplyPatch(patch AddressPatch) error {
	if a.IsDeleted() {
		return fmt.Errorf("%w: address %q", ErrDeletedResource, a.AddressID)
	}
	if patch.UpdatedAt.IsZero() {
		return ValidationError{Fields: []FieldError{{Field: "updated_at", Message: "is required"}}}
	}
	if trim(patch.UpdatedBy) == "" {
		return ValidationError{Fields: []FieldError{{Field: "updated_by", Message: "is required"}}}
	}

	next := *a
	if patch.Name != nil {
		next.Name = cleanText(*patch.Name)
	}
	if patch.Phone != nil {
		next.Phone = cleanOptionalPhone(patch.Phone)
	}
	if patch.Line1 != nil {
		next.Line1 = cleanText(*patch.Line1)
	}
	if patch.Line2 != nil {
		next.Line2 = cleanOptionalText(patch.Line2)
	}
	if patch.City != nil {
		next.City = cleanText(*patch.City)
	}
	if patch.State != nil {
		next.State = cleanText(*patch.State)
	}
	if patch.PostalCode != nil {
		next.PostalCode = trim(*patch.PostalCode)
	}
	if patch.Country != nil {
		next.Country = cleanCountry(*patch.Country)
	}
	if patch.IsDefault != nil {
		next.IsDefault = *patch.IsDefault
	}
	next.UpdatedBy = trim(patch.UpdatedBy)
	next.UpdatedAt = patch.UpdatedAt.UTC()

	if err := next.Validate(); err != nil {
		return err
	}

	*a = next
	return nil
}

func (a *Address) MarkDeleted(actorID string, at time.Time) error {
	actorID = trim(actorID)
	if actorID == "" {
		return ValidationError{Fields: []FieldError{{Field: "deleted_by", Message: "is required"}}}
	}
	if at.IsZero() {
		return ValidationError{Fields: []FieldError{{Field: "deleted_at", Message: "is required"}}}
	}
	if a.IsDeleted() {
		return nil
	}

	deletedAt := at.UTC()
	a.Status = AddressStatusDeleted
	a.DeletedBy = &actorID
	a.DeletedAt = &deletedAt
	a.UpdatedBy = actorID
	a.UpdatedAt = deletedAt
	a.IsDefault = false
	return a.Validate()
}
