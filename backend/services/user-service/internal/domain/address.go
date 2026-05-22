package domain

import (
	"fmt"
	"time"
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
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
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
	UpdatedAt  time.Time
}

func NewAddress(params NewAddressParams) (Address, error) {
	createdAt := params.CreatedAt.UTC()
	address := Address{
		AddressID:  trim(params.AddressID),
		UserID:     trim(params.UserID),
		Name:       trim(params.Name),
		Phone:      cleanOptional(params.Phone),
		Line1:      trim(params.Line1),
		Line2:      cleanOptional(params.Line2),
		City:       trim(params.City),
		State:      trim(params.State),
		PostalCode: trim(params.PostalCode),
		Country:    trim(params.Country),
		IsDefault:  params.IsDefault,
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
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
	validateRequiredString(&v, "name", a.Name, maxNameLength)
	validateOptionalPhone(&v, "phone", a.Phone)
	validateRequiredString(&v, "line1", a.Line1, maxAddressLineLength)
	validateOptionalString(&v, "line2", a.Line2, maxAddressLineLength)
	validateRequiredString(&v, "city", a.City, maxCityStateLength)
	validateRequiredString(&v, "state", a.State, maxCityStateLength)
	validateRequiredString(&v, "postal_code", a.PostalCode, maxPostalCodeLength)
	validateRequiredString(&v, "country", a.Country, maxCountryLength)
	validateTimestamp(&v, "created_at", a.CreatedAt)
	validateTimestamp(&v, "updated_at", a.UpdatedAt)
	if !a.CreatedAt.IsZero() && !a.UpdatedAt.IsZero() && a.UpdatedAt.Before(a.CreatedAt) {
		v.add("updated_at", "cannot be before created_at")
	}
	if a.DeletedAt != nil && a.DeletedAt.Before(a.CreatedAt) {
		v.add("deleted_at", "cannot be before created_at")
	}
	if a.DeletedAt != nil && a.IsDefault {
		v.add("is_default", "deleted address cannot be default")
	}

	return v.err()
}

func (a Address) BelongsTo(userID string) bool {
	return a.UserID == trim(userID)
}

func (a Address) IsDeleted() bool {
	return a.DeletedAt != nil
}

func (a *Address) ApplyPatch(patch AddressPatch) error {
	if a.IsDeleted() {
		return fmt.Errorf("%w: address %q", ErrDeletedResource, a.AddressID)
	}
	if patch.UpdatedAt.IsZero() {
		return ValidationError{Fields: []FieldError{{Field: "updated_at", Message: "is required"}}}
	}

	next := *a
	if patch.Name != nil {
		next.Name = trim(*patch.Name)
	}
	if patch.Phone != nil {
		next.Phone = cleanOptional(patch.Phone)
	}
	if patch.Line1 != nil {
		next.Line1 = trim(*patch.Line1)
	}
	if patch.Line2 != nil {
		next.Line2 = cleanOptional(patch.Line2)
	}
	if patch.City != nil {
		next.City = trim(*patch.City)
	}
	if patch.State != nil {
		next.State = trim(*patch.State)
	}
	if patch.PostalCode != nil {
		next.PostalCode = trim(*patch.PostalCode)
	}
	if patch.Country != nil {
		next.Country = trim(*patch.Country)
	}
	if patch.IsDefault != nil {
		next.IsDefault = *patch.IsDefault
	}
	next.UpdatedAt = patch.UpdatedAt.UTC()

	if err := next.Validate(); err != nil {
		return err
	}

	*a = next
	return nil
}

func (a *Address) MarkDeleted(at time.Time) error {
	if at.IsZero() {
		return ValidationError{Fields: []FieldError{{Field: "deleted_at", Message: "is required"}}}
	}
	if a.IsDeleted() {
		return nil
	}

	deletedAt := at.UTC()
	a.DeletedAt = &deletedAt
	a.UpdatedAt = deletedAt
	a.IsDefault = false
	return a.Validate()
}
