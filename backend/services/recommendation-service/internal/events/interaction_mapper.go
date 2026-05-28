package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

const (
	defaultMaxIdentifierLength    = 128
	defaultMaxMetadataFields      = 16
	defaultMaxMetadataValueLength = 512
)

type InteractionMapperConfig struct {
	MaxIdentifierLength    int
	MaxMetadataFields      int
	MaxMetadataValueLength int
	SupportedVersion       int
}

type InteractionMapper struct {
	maxIdentifierLength    int
	maxMetadataFields      int
	maxMetadataValueLength int
	supportedVersion       int
}

type metadataEntry struct {
	key   string
	value any
}

func NewInteractionMapper(cfg InteractionMapperConfig) (*InteractionMapper, error) {
	if cfg.MaxIdentifierLength <= 0 {
		cfg.MaxIdentifierLength = defaultMaxIdentifierLength
	}
	if cfg.MaxMetadataFields <= 0 {
		cfg.MaxMetadataFields = defaultMaxMetadataFields
	}
	if cfg.MaxMetadataValueLength <= 0 {
		cfg.MaxMetadataValueLength = defaultMaxMetadataValueLength
	}
	if cfg.SupportedVersion <= 0 {
		cfg.SupportedVersion = DefaultSupportedEventVersion
	}
	return &InteractionMapper{
		maxIdentifierLength:    cfg.MaxIdentifierLength,
		maxMetadataFields:      cfg.MaxMetadataFields,
		maxMetadataValueLength: cfg.MaxMetadataValueLength,
		supportedVersion:       cfg.SupportedVersion,
	}, nil
}

func (m *InteractionMapper) Map(envelope Envelope, receivedAt time.Time) ([]domain.UserInteraction, error) {
	if m == nil {
		return nil, fmt.Errorf("%w: interaction mapper is not initialized", domain.ErrInvalidInteraction)
	}
	if receivedAt.IsZero() {
		receivedAt = time.Now().UTC()
	}
	if err := envelope.Validate(m.supportedVersion); err != nil {
		return nil, err
	}

	switch envelope.EventType {
	case domain.EventProductViewed,
		domain.EventCartItemAdded,
		domain.EventWishlistItemAdded,
		domain.EventWishlistItemRemoved:
		return m.mapProductInteraction(envelope, receivedAt)
	case domain.EventOrderPaid, domain.EventPurchaseCompleted:
		return m.mapPurchaseInteraction(envelope, receivedAt)
	default:
		return nil, fmt.Errorf("%w: %s", domain.ErrUnsupportedEventType, envelope.EventType)
	}
}

func (m *InteractionMapper) mapProductInteraction(envelope Envelope, receivedAt time.Time) ([]domain.UserInteraction, error) {
	var payload ProductInteractionPayload
	if err := decodePayload(envelope.Payload, &payload); err != nil {
		return nil, err
	}

	quantity := payload.Quantity
	if envelope.EventType == domain.EventProductViewed ||
		envelope.EventType == domain.EventWishlistItemAdded ||
		envelope.EventType == domain.EventWishlistItemRemoved {
		quantity = 1
	}

	interaction, err := m.buildInteraction(envelope, receivedAt, productSignal{
		UserID:      payload.UserID,
		AnonymousID: payload.AnonymousID,
		SessionID:   payload.SessionID,
		ProductID:   payload.ProductID,
		VariantID:   payload.VariantID,
		CategoryID:  payload.CategoryID,
		SellerID:    payload.SellerID,
		BrandID:     payload.BrandID,
		Quantity:    quantity,
		Metadata: m.safeMetadata(payload.Metadata,
			metadataEntry{key: "page", value: payload.Page},
			metadataEntry{key: "cart_id", value: payload.CartID},
			metadataEntry{key: "wishlist_id", value: payload.WishlistID},
		),
	})
	if err != nil {
		return nil, err
	}
	return []domain.UserInteraction{interaction}, nil
}

func (m *InteractionMapper) mapPurchaseInteraction(envelope Envelope, receivedAt time.Time) ([]domain.UserInteraction, error) {
	var payload PurchasePayload
	if err := decodePayload(envelope.Payload, &payload); err != nil {
		return nil, err
	}
	if err := m.validateIdentity(payload.UserID, payload.AnonymousID); err != nil {
		return nil, err
	}
	if len(payload.Items) == 0 {
		return nil, fmt.Errorf("%w: purchase items are required", domain.ErrInvalidInteraction)
	}

	interactions := make([]domain.UserInteraction, 0, len(payload.Items))
	for index, item := range payload.Items {
		metadata := m.safeMetadata(payload.Metadata,
			metadataEntry{key: "order_id", value: payload.OrderID},
			metadataEntry{key: "item_index", value: index},
			metadataEntry{key: "unit_price_amount", value: item.UnitPriceAmount},
			metadataEntry{key: "currency", value: item.Currency},
		)
		if metadata == nil {
			metadata = make(map[string]any)
		}
		for key, value := range m.safeMetadata(item.Metadata) {
			if len(metadata) >= m.maxMetadataFields {
				break
			}
			metadata["item_"+key] = value
		}

		interaction, err := m.buildInteraction(envelope, receivedAt, productSignal{
			UserID:      payload.UserID,
			AnonymousID: payload.AnonymousID,
			SessionID:   payload.SessionID,
			ProductID:   item.ProductID,
			VariantID:   item.VariantID,
			CategoryID:  item.CategoryID,
			SellerID:    item.SellerID,
			BrandID:     item.BrandID,
			Quantity:    item.Quantity,
			Metadata:    metadata,
		})
		if err != nil {
			return nil, fmt.Errorf("%w: item %d: %v", domain.ErrInvalidInteraction, index, err)
		}
		interactions = append(interactions, interaction)
	}
	return interactions, nil
}

type productSignal struct {
	UserID      string
	AnonymousID string
	SessionID   string
	ProductID   string
	VariantID   string
	CategoryID  string
	SellerID    string
	BrandID     string
	Quantity    int
	Metadata    map[string]any
}

func (m *InteractionMapper) buildInteraction(envelope Envelope, receivedAt time.Time, signal productSignal) (domain.UserInteraction, error) {
	userID := normalizeIdentifier(signal.UserID)
	anonymousID := normalizeIdentifier(signal.AnonymousID)
	if err := m.validateIdentity(userID, anonymousID); err != nil {
		return domain.UserInteraction{}, err
	}

	productID := normalizeIdentifier(signal.ProductID)
	if err := m.validateIdentifier("product_id", productID, true); err != nil {
		return domain.UserInteraction{}, err
	}
	variantID := normalizeIdentifier(signal.VariantID)
	if err := m.validateIdentifier("variant_id", variantID, false); err != nil {
		return domain.UserInteraction{}, err
	}
	categoryID := normalizeIdentifier(signal.CategoryID)
	if err := m.validateIdentifier("category_id", categoryID, false); err != nil {
		return domain.UserInteraction{}, err
	}
	if err := validateFeatureDimensionKey("category_id", categoryID); err != nil {
		return domain.UserInteraction{}, err
	}
	sellerID := normalizeIdentifier(signal.SellerID)
	if err := m.validateIdentifier("seller_id", sellerID, false); err != nil {
		return domain.UserInteraction{}, err
	}
	if err := validateFeatureDimensionKey("seller_id", sellerID); err != nil {
		return domain.UserInteraction{}, err
	}
	brandID := normalizeIdentifier(signal.BrandID)
	if err := m.validateIdentifier("brand_id", brandID, false); err != nil {
		return domain.UserInteraction{}, err
	}
	if err := validateFeatureDimensionKey("brand_id", brandID); err != nil {
		return domain.UserInteraction{}, err
	}
	sessionID := normalizeIdentifier(signal.SessionID)
	if err := m.validateIdentifier("session_id", sessionID, false); err != nil {
		return domain.UserInteraction{}, err
	}

	normalizedType, ok := envelope.EventType.NormalizedInteractionType()
	if !ok {
		return domain.UserInteraction{}, fmt.Errorf("%w: %s", domain.ErrUnsupportedEventType, envelope.EventType)
	}
	weight, ok := envelope.EventType.InteractionWeight()
	if !ok {
		return domain.UserInteraction{}, fmt.Errorf("%w: %s", domain.ErrUnsupportedEventType, envelope.EventType)
	}
	if signal.Quantity <= 0 {
		return domain.UserInteraction{}, fmt.Errorf("%w: quantity must be greater than zero", domain.ErrInvalidInteraction)
	}

	dedupeKey := domain.NewInteractionDedupeKey(envelope.EventID, normalizedType, productID, variantID)
	interaction := domain.UserInteraction{
		ID:                  domain.NewInteractionID(dedupeKey),
		DedupeKey:           dedupeKey,
		EventID:             envelope.EventID,
		SourceEventType:     envelope.EventType,
		NormalizedEventType: normalizedType,
		Version:             envelope.Version,
		Producer:            envelope.Producer,
		TraceID:             envelope.TraceID,
		UserID:              userID,
		AnonymousID:         anonymousID,
		SessionID:           sessionID,
		ProductID:           productID,
		VariantID:           variantID,
		CategoryID:          categoryID,
		SellerID:            sellerID,
		BrandID:             brandID,
		Weight:              weight,
		Quantity:            signal.Quantity,
		Metadata:            signal.Metadata,
		OccurredAt:          envelope.OccurredAt.UTC(),
		ReceivedAt:          receivedAt.UTC(),
	}
	if err := interaction.Validate(); err != nil {
		return domain.UserInteraction{}, err
	}
	return interaction, nil
}

func decodePayload(raw []byte, dst any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("%w: invalid payload: %v", domain.ErrInvalidInteraction, err)
	}
	var extra struct{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("%w: payload must contain a single json object", domain.ErrInvalidInteraction)
	}
	return nil
}

func (m *InteractionMapper) validateIdentity(userID string, anonymousID string) error {
	if err := m.validateIdentifier("user_id", userID, false); err != nil {
		return err
	}
	if err := m.validateIdentifier("anonymous_id", anonymousID, false); err != nil {
		return err
	}
	if userID == "" && anonymousID == "" {
		return fmt.Errorf("%w: user_id or anonymous_id is required", domain.ErrInvalidInteraction)
	}
	return nil
}

func (m *InteractionMapper) validateIdentifier(field string, value string, required bool) error {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			return fmt.Errorf("%w: %s is required", domain.ErrInvalidInteraction, field)
		}
		return nil
	}
	if len(value) > m.maxIdentifierLength {
		return fmt.Errorf("%w: %s is too long", domain.ErrInvalidInteraction, field)
	}
	if strings.ContainsAny(value, "\x00\r\n\t") {
		return fmt.Errorf("%w: %s contains unsupported whitespace/control characters", domain.ErrInvalidInteraction, field)
	}
	return nil
}

func normalizeIdentifier(value string) string {
	return strings.TrimSpace(value)
}

func validateFeatureDimensionKey(field string, value string) error {
	if strings.Contains(value, ".") || strings.HasPrefix(value, "$") {
		return fmt.Errorf("%w: %s cannot contain MongoDB field path characters", domain.ErrInvalidInteraction, field)
	}
	return nil
}

func (m *InteractionMapper) safeMetadata(base map[string]any, entries ...metadataEntry) map[string]any {
	metadata := make(map[string]any)
	add := func(key string, value any) {
		if len(metadata) >= m.maxMetadataFields {
			return
		}
		key = strings.TrimSpace(key)
		if key == "" || sensitiveMetadataKey(key) {
			return
		}
		safeValue, ok := m.safeMetadataValue(value)
		if !ok {
			return
		}
		metadata[key] = safeValue
	}

	for key, value := range base {
		add(key, value)
	}
	for _, entry := range entries {
		add(entry.key, entry.value)
	}
	if len(metadata) == 0 {
		return nil
	}
	return metadata
}

func (m *InteractionMapper) safeMetadataValue(value any) (any, bool) {
	switch v := value.(type) {
	case nil:
		return nil, false
	case string:
		v = strings.TrimSpace(v)
		if v == "" || len(v) > m.maxMetadataValueLength {
			return nil, false
		}
		return v, true
	case bool:
		return v, true
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, false
		}
		return v, true
	case float32:
		f := float64(v)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, false
		}
		return v, true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return v, true
	default:
		return nil, false
	}
}

func sensitiveMetadataKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(key), "-", "_"))
	sensitiveTokens := []string{
		"email",
		"phone",
		"full_name",
		"name",
		"address",
		"ip",
		"ip_address",
		"raw_ip",
		"token",
		"password",
		"secret",
		"card",
		"card_number",
		"cvv",
	}
	for _, token := range sensitiveTokens {
		if normalized == token || strings.Contains(normalized, "_"+token) || strings.Contains(normalized, token+"_") {
			return true
		}
	}
	return false
}
