package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	CurrentHeatmapSchemaVersion = 1

	DefaultHeatmapClickBucketSize = 5
	ScrollDepthBucketQuarter      = 25
	ScrollDepthBucketHalf         = 50
	ScrollDepthBucketThreeFourth  = 75
	ScrollDepthBucketNearFull     = 90
	ScrollDepthBucketFull         = 100
)

type HeatmapType string

const (
	HeatmapTypeClick  HeatmapType = "click"
	HeatmapTypeScroll HeatmapType = "scroll"
)

func (t HeatmapType) Valid() bool {
	switch t {
	case HeatmapTypeClick, HeatmapTypeScroll:
		return true
	default:
		return false
	}
}

type HeatmapPoint struct {
	ID             string         `json:"id,omitempty" bson:"_id,omitempty"`
	HeatmapType    HeatmapType    `json:"heatmap_type" bson:"heatmap_type"`
	Path           string         `json:"path" bson:"path"`
	NormalizedPath string         `json:"normalized_path" bson:"normalized_path"`
	DeviceType     DeviceType     `json:"device_type" bson:"device_type"`
	ViewportBucket string         `json:"viewport_bucket" bson:"viewport_bucket"`
	Day            string         `json:"day" bson:"day"`
	XBucket        *int           `json:"x_bucket,omitempty" bson:"x_bucket,omitempty"`
	YBucket        *int           `json:"y_bucket,omitempty" bson:"y_bucket,omitempty"`
	DepthBucket    *int           `json:"depth_bucket,omitempty" bson:"depth_bucket,omitempty"`
	X              int            `json:"x" bson:"x"`
	Y              int            `json:"y" bson:"y"`
	Weight         int            `json:"weight" bson:"weight"`
	UniqueSessions int            `json:"unique_sessions" bson:"unique_sessions"`
	SampleEvents   int            `json:"sample_events" bson:"sample_events"`
	SchemaVersion  int            `json:"schema_version" bson:"schema_version"`
	FirstSeenAt    time.Time      `json:"first_seen_at" bson:"first_seen_at"`
	LastSeenAt     time.Time      `json:"last_seen_at" bson:"last_seen_at"`
	CreatedAt      time.Time      `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt      time.Time      `json:"updated_at" bson:"updated_at"`
	RetainUntil    *time.Time     `json:"retain_until,omitempty" bson:"retain_until,omitempty"`
	RetentionClass RetentionClass `json:"retention_class,omitempty" bson:"retention_class,omitempty"`
}

type HeatmapFilter struct {
	HeatmapType    HeatmapType
	Path           string
	DeviceType     DeviceType
	ViewportBucket string
	FromDay        string
	ToDay          string
}

type HeatmapAggregationCheckpoint struct {
	WorkerName      string    `json:"worker_name" bson:"worker_name"`
	LastProcessedAt time.Time `json:"last_processed_at" bson:"last_processed_at"`
	UpdatedAt       time.Time `json:"updated_at" bson:"updated_at"`
}

type HeatmapEventCursor struct {
	OccurredAt time.Time
	EventID    string
}

type ClickProperties struct {
	ElementID      string  `json:"element_id,omitempty" bson:"element_id,omitempty"`
	X              float64 `json:"x" bson:"x"`
	Y              float64 `json:"y" bson:"y"`
	ViewportWidth  float64 `json:"viewport_width" bson:"viewport_width"`
	ViewportHeight float64 `json:"viewport_height" bson:"viewport_height"`
}

type ScrollProperties struct {
	ViewportWidth  float64 `json:"viewport_width,omitempty" bson:"viewport_width,omitempty"`
	DepthPercent   float64 `json:"depth_percent" bson:"depth_percent"`
	ViewportHeight float64 `json:"viewport_height,omitempty" bson:"viewport_height,omitempty"`
	DocumentHeight float64 `json:"document_height,omitempty" bson:"document_height,omitempty"`
}

type HeatmapValidationError struct {
	Violations []FieldViolation
}

func (e HeatmapValidationError) Error() string {
	if len(e.Violations) == 0 {
		return ErrInvalidHeatmap.Error()
	}
	parts := make([]string, 0, len(e.Violations))
	for _, violation := range e.Violations {
		parts = append(parts, fmt.Sprintf("%s: %s", violation.Field, violation.Message))
	}
	return ErrInvalidHeatmap.Error() + ": " + strings.Join(parts, "; ")
}

func (e HeatmapValidationError) Is(target error) bool {
	return target == ErrInvalidHeatmap
}

func HeatmapPointFromEvent(event SessionEvent, now time.Time, clickBucketSize int) (HeatmapPoint, error) {
	normalized := event.Normalize()
	if normalized.Path == nil || strings.TrimSpace(*normalized.Path) == "" {
		return HeatmapPoint{}, HeatmapValidationError{Violations: []FieldViolation{{Field: "path", Message: "is required"}}}
	}
	if clickBucketSize <= 0 {
		clickBucketSize = DefaultHeatmapClickBucketSize
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	path := strings.TrimSpace(*normalized.Path)
	deviceType := heatmapDeviceType(normalized)
	day := normalized.OccurredAt.UTC().Format("2006-01-02")
	base := HeatmapPoint{
		Path:           path,
		NormalizedPath: NormalizeHeatmapPath(path),
		DeviceType:     deviceType,
		Day:            day,
		SchemaVersion:  CurrentHeatmapSchemaVersion,
		FirstSeenAt:    normalized.OccurredAt.UTC(),
		LastSeenAt:     normalized.OccurredAt.UTC(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	switch normalized.EventType {
	case EventClick:
		props, err := ClickPropertiesFromMap(normalized.Properties)
		if err != nil {
			return HeatmapPoint{}, err
		}
		if err := props.Validate(); err != nil {
			return HeatmapPoint{}, err
		}
		xPercent := NormalizeCoordinatePercent(props.X, props.ViewportWidth)
		yPercent := NormalizeCoordinatePercent(props.Y, props.ViewportHeight)
		xBucket := BucketPercent(xPercent, clickBucketSize)
		yBucket := BucketPercent(yPercent, clickBucketSize)
		base.HeatmapType = HeatmapTypeClick
		base.ViewportBucket = ViewportBucket(deviceType, props.ViewportWidth)
		base.XBucket = intPtr(xBucket)
		base.YBucket = intPtr(yBucket)
		base.X = xBucket
		base.Y = yBucket
	case EventScroll:
		props, err := ScrollPropertiesFromMap(normalized.Properties)
		if err != nil {
			return HeatmapPoint{}, err
		}
		if err := props.Validate(); err != nil {
			return HeatmapPoint{}, err
		}
		depthBucket := BucketScrollDepth(props.DepthPercent)
		base.HeatmapType = HeatmapTypeScroll
		base.ViewportBucket = ViewportBucket(deviceType, heatmapViewportWidth(normalized, props))
		base.DepthBucket = intPtr(depthBucket)
		base.X = 50
		base.Y = depthBucket
	default:
		return HeatmapPoint{}, HeatmapValidationError{Violations: []FieldViolation{{Field: "event_type", Message: "must be click or scroll"}}}
	}
	base.ID = base.DeterministicID()
	return base.Normalize(), nil
}

func (p HeatmapPoint) Normalize() HeatmapPoint {
	out := p
	out.ID = strings.TrimSpace(out.ID)
	out.HeatmapType = HeatmapType(strings.TrimSpace(string(out.HeatmapType)))
	out.Path = strings.TrimSpace(out.Path)
	out.NormalizedPath = strings.TrimSpace(out.NormalizedPath)
	if out.NormalizedPath == "" {
		out.NormalizedPath = NormalizeHeatmapPath(out.Path)
	}
	out.DeviceType = DeviceType(strings.TrimSpace(string(out.DeviceType)))
	if out.DeviceType == "" {
		out.DeviceType = DeviceTypeUnknown
	}
	out.ViewportBucket = strings.TrimSpace(out.ViewportBucket)
	if out.ViewportBucket == "" {
		out.ViewportBucket = "unknown"
	}
	out.Day = strings.TrimSpace(out.Day)
	out.XBucket = normalizeIntPtr(out.XBucket)
	out.YBucket = normalizeIntPtr(out.YBucket)
	out.DepthBucket = normalizeIntPtr(out.DepthBucket)
	out.FirstSeenAt = normalizeTime(out.FirstSeenAt)
	out.LastSeenAt = normalizeTime(out.LastSeenAt)
	out.CreatedAt = normalizeTime(out.CreatedAt)
	out.UpdatedAt = normalizeTime(out.UpdatedAt)
	out.RetainUntil = normalizeTimePtr(out.RetainUntil)
	out.RetentionClass = RetentionClass(strings.TrimSpace(string(out.RetentionClass)))
	if out.SchemaVersion == 0 {
		out.SchemaVersion = CurrentHeatmapSchemaVersion
	}
	if out.ID == "" && out.HeatmapType.Valid() && out.Path != "" && out.Day != "" {
		out.ID = out.DeterministicID()
	}
	return out
}

func (p HeatmapPoint) Validate() error {
	point := p.Normalize()
	violations := make([]FieldViolation, 0)
	if !point.HeatmapType.Valid() {
		violations = append(violations, FieldViolation{Field: "heatmap_type", Message: "must be click or scroll"})
	}
	validateHeatmapPath(&violations, "path", point.Path, true, DefaultMaxPageLength)
	validateHeatmapPath(&violations, "normalized_path", point.NormalizedPath, true, DefaultMaxPageLength)
	if !point.DeviceType.Valid() {
		violations = append(violations, FieldViolation{Field: "device_type", Message: "must be a supported device type"})
	}
	if point.ViewportBucket == "" {
		violations = append(violations, FieldViolation{Field: "viewport_bucket", Message: "is required"})
	}
	if !validDay(point.Day) {
		violations = append(violations, FieldViolation{Field: "day", Message: "must use YYYY-MM-DD format"})
	}
	if point.SchemaVersion != CurrentHeatmapSchemaVersion {
		violations = append(violations, FieldViolation{Field: "schema_version", Message: "must match current heatmap schema version"})
	}
	if point.Weight < 0 || point.UniqueSessions < 0 || point.SampleEvents < 0 {
		violations = append(violations, FieldViolation{Field: "counts", Message: "cannot be negative"})
	}
	if point.X < 0 || point.X > 100 || point.Y < 0 || point.Y > 100 {
		violations = append(violations, FieldViolation{Field: "coordinates", Message: "must be between 0 and 100"})
	}
	switch point.HeatmapType {
	case HeatmapTypeClick:
		if point.XBucket == nil || point.YBucket == nil {
			violations = append(violations, FieldViolation{Field: "click_buckets", Message: "x_bucket and y_bucket are required for click heatmaps"})
		}
		if point.DepthBucket != nil {
			violations = append(violations, FieldViolation{Field: "depth_bucket", Message: "must be empty for click heatmaps"})
		}
	case HeatmapTypeScroll:
		if point.DepthBucket == nil {
			violations = append(violations, FieldViolation{Field: "depth_bucket", Message: "is required for scroll heatmaps"})
		}
		if point.XBucket != nil || point.YBucket != nil {
			violations = append(violations, FieldViolation{Field: "click_buckets", Message: "must be empty for scroll heatmaps"})
		}
	}
	if point.FirstSeenAt.IsZero() {
		violations = append(violations, FieldViolation{Field: "first_seen_at", Message: "is required"})
	}
	if point.LastSeenAt.IsZero() {
		violations = append(violations, FieldViolation{Field: "last_seen_at", Message: "is required"})
	}
	if !point.FirstSeenAt.IsZero() && !point.LastSeenAt.IsZero() && point.LastSeenAt.Before(point.FirstSeenAt) {
		violations = append(violations, FieldViolation{Field: "last_seen_at", Message: "cannot be before first_seen_at"})
	}
	if point.UpdatedAt.IsZero() {
		violations = append(violations, FieldViolation{Field: "updated_at", Message: "is required"})
	}
	validateOptionalRetentionMetadata(&violations, point.RetentionClass, false, nil, nil)
	if len(violations) > 0 {
		return HeatmapValidationError{Violations: violations}
	}
	return nil
}

func (p HeatmapPoint) BucketKey() string {
	point := p
	point.HeatmapType = HeatmapType(strings.TrimSpace(string(point.HeatmapType)))
	point.Path = strings.TrimSpace(point.Path)
	point.DeviceType = DeviceType(strings.TrimSpace(string(point.DeviceType)))
	point.ViewportBucket = strings.TrimSpace(point.ViewportBucket)
	point.Day = strings.TrimSpace(point.Day)
	parts := []string{
		string(point.HeatmapType),
		point.Path,
		string(point.DeviceType),
		point.ViewportBucket,
		point.Day,
	}
	if point.HeatmapType == HeatmapTypeClick {
		parts = append(parts, intPtrString(point.XBucket), intPtrString(point.YBucket))
	} else {
		parts = append(parts, intPtrString(point.DepthBucket))
	}
	return strings.Join(parts, "|")
}

func (p HeatmapPoint) DeterministicID() string {
	sum := sha256.Sum256([]byte(p.BucketKey()))
	return "hm_" + hex.EncodeToString(sum[:16])
}

func (p HeatmapPoint) SessionMarkerID(sessionID string) string {
	sum := sha256.Sum256([]byte(p.BucketKey() + "|" + strings.TrimSpace(sessionID)))
	return "hms_" + hex.EncodeToString(sum[:16])
}

func (f HeatmapFilter) Normalize() HeatmapFilter {
	out := f
	out.HeatmapType = HeatmapType(strings.TrimSpace(string(out.HeatmapType)))
	out.Path = strings.TrimSpace(out.Path)
	out.DeviceType = DeviceType(strings.TrimSpace(string(out.DeviceType)))
	out.ViewportBucket = strings.TrimSpace(out.ViewportBucket)
	out.FromDay = strings.TrimSpace(out.FromDay)
	out.ToDay = strings.TrimSpace(out.ToDay)
	return out
}

func (f HeatmapFilter) Validate() error {
	filter := f.Normalize()
	violations := make([]FieldViolation, 0)
	if !filter.HeatmapType.Valid() {
		violations = append(violations, FieldViolation{Field: "heatmap_type", Message: "must be click or scroll"})
	}
	validateHeatmapPath(&violations, "path", filter.Path, true, DefaultMaxPageLength)
	if !filter.DeviceType.Valid() {
		violations = append(violations, FieldViolation{Field: "device_type", Message: "must be a supported device type"})
	}
	if !validDay(filter.FromDay) {
		violations = append(violations, FieldViolation{Field: "from", Message: "must use YYYY-MM-DD format"})
	}
	if !validDay(filter.ToDay) {
		violations = append(violations, FieldViolation{Field: "to", Message: "must use YYYY-MM-DD format"})
	}
	if validDay(filter.FromDay) && validDay(filter.ToDay) && strings.Compare(filter.FromDay, filter.ToDay) > 0 {
		violations = append(violations, FieldViolation{Field: "from", Message: "cannot be after to"})
	}
	if len(violations) > 0 {
		return HeatmapValidationError{Violations: violations}
	}
	return nil
}

func (p ClickProperties) Validate() error {
	violations := make([]FieldViolation, 0)
	if p.ViewportWidth <= 0 {
		violations = append(violations, FieldViolation{Field: "viewport_width", Message: "must be greater than zero"})
	}
	if p.ViewportHeight <= 0 {
		violations = append(violations, FieldViolation{Field: "viewport_height", Message: "must be greater than zero"})
	}
	if p.X < 0 || p.Y < 0 {
		violations = append(violations, FieldViolation{Field: "coordinates", Message: "cannot be negative"})
	}
	if p.ViewportWidth > 0 && p.X > p.ViewportWidth {
		violations = append(violations, FieldViolation{Field: "x", Message: "cannot be outside viewport"})
	}
	if p.ViewportHeight > 0 && p.Y > p.ViewportHeight {
		violations = append(violations, FieldViolation{Field: "y", Message: "cannot be outside viewport"})
	}
	if len(violations) > 0 {
		return HeatmapValidationError{Violations: violations}
	}
	return nil
}

func (p ScrollProperties) Validate() error {
	if p.DepthPercent < 0 || p.DepthPercent > 100 {
		return HeatmapValidationError{Violations: []FieldViolation{{Field: "depth_percent", Message: "must be between 0 and 100"}}}
	}
	return nil
}

func ClickPropertiesFromMap(properties map[string]any) (ClickProperties, error) {
	props := ClickProperties{
		ElementID:      stringProperty(properties, "element_id"),
		X:              numericProperty(properties, "x"),
		Y:              numericProperty(properties, "y"),
		ViewportWidth:  numericProperty(properties, "viewport_width"),
		ViewportHeight: numericProperty(properties, "viewport_height"),
	}
	missing := missingNumericKeys(properties, "x", "y", "viewport_width", "viewport_height")
	if len(missing) > 0 {
		return ClickProperties{}, HeatmapValidationError{Violations: []FieldViolation{{Field: strings.Join(missing, ","), Message: "is required"}}}
	}
	return props, nil
}

func ScrollPropertiesFromMap(properties map[string]any) (ScrollProperties, error) {
	props := ScrollProperties{
		ViewportWidth:  numericProperty(properties, "viewport_width"),
		DepthPercent:   numericProperty(properties, "depth_percent"),
		ViewportHeight: numericProperty(properties, "viewport_height"),
		DocumentHeight: numericProperty(properties, "document_height"),
	}
	missing := missingNumericKeys(properties, "depth_percent")
	if len(missing) > 0 {
		return ScrollProperties{}, HeatmapValidationError{Violations: []FieldViolation{{Field: "depth_percent", Message: "is required"}}}
	}
	return props, nil
}

func NormalizeCoordinatePercent(value float64, total float64) float64 {
	if total <= 0 {
		return 0
	}
	percent := value / total * 100
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

func BucketPercent(value float64, bucketSize int) int {
	if bucketSize <= 0 {
		bucketSize = DefaultHeatmapClickBucketSize
	}
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	bucket := int(math.Round(value/float64(bucketSize))) * bucketSize
	if bucket < 0 {
		return 0
	}
	if bucket > 100 {
		return 100
	}
	return bucket
}

func BucketScrollDepth(depth float64) int {
	switch {
	case depth >= ScrollDepthBucketFull:
		return ScrollDepthBucketFull
	case depth >= ScrollDepthBucketNearFull:
		return ScrollDepthBucketNearFull
	case depth >= ScrollDepthBucketThreeFourth:
		return ScrollDepthBucketThreeFourth
	case depth >= ScrollDepthBucketHalf:
		return ScrollDepthBucketHalf
	case depth >= ScrollDepthBucketQuarter:
		return ScrollDepthBucketQuarter
	default:
		return 0
	}
}

func ViewportBucket(deviceType DeviceType, width float64) string {
	switch deviceType {
	case DeviceTypeMobile:
		if width <= 360 {
			return "mobile_0_360"
		}
		if width <= 480 {
			return "mobile_360_480"
		}
		return "mobile_480_plus"
	case DeviceTypeTablet:
		if width <= 1024 {
			return "tablet_768_1024"
		}
		return "tablet_1024_plus"
	case DeviceTypeDesktop:
		if width <= 1440 {
			return "desktop_1024_1440"
		}
		return "desktop_1440_plus"
	case DeviceTypeBot:
		return "bot"
	default:
		return "unknown"
	}
}

func NormalizeHeatmapPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" || trimmed == "/" {
		return trimmed
	}
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) == 2 {
		switch parts[0] {
		case "products":
			return "/products/:product_id"
		case "categories":
			return "/categories/:category_id"
		case "sellers":
			return "/sellers/:seller_id"
		}
	}
	for i, part := range parts {
		if dynamicPathSegment(part) {
			parts[i] = ":id"
		}
	}
	return "/" + strings.Join(parts, "/")
}

func NormalizeHeatmapType(value string) HeatmapType {
	value = strings.TrimSpace(value)
	if value == "" {
		return HeatmapTypeClick
	}
	return HeatmapType(value)
}

func DayString(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

func validDay(value string) bool {
	if value == "" {
		return false
	}
	parsed, err := time.Parse("2006-01-02", value)
	return err == nil && parsed.Format("2006-01-02") == value
}

func validateHeatmapPath(violations *[]FieldViolation, field string, value string, required bool, maxLength int) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			addViolation(violations, field, "is required")
		}
		return
	}
	if len(value) > maxLength {
		addViolation(violations, field, "is too long")
		return
	}
	if !strings.HasPrefix(value, "/") {
		addViolation(violations, field, "must start with /")
	}
}

func heatmapDeviceType(event SessionEvent) DeviceType {
	if event.Device != nil {
		device := event.Device.Normalize()
		if device.Type.Valid() {
			return device.Type
		}
	}
	return DeviceTypeUnknown
}

func heatmapViewportWidth(event SessionEvent, props ScrollProperties) float64 {
	if props.ViewportWidth > 0 {
		return props.ViewportWidth
	}
	if event.Client != nil && event.Client.ViewportWidth > 0 {
		return float64(event.Client.ViewportWidth)
	}
	return props.ViewportHeight
}

func numericProperty(properties map[string]any, key string) float64 {
	value, _ := numericPropertyOK(properties, key)
	return value
}

func numericPropertyOK(properties map[string]any, key string) (float64, bool) {
	value, ok := properties[key]
	if !ok {
		return 0, false
	}
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func missingNumericKeys(properties map[string]any, keys ...string) []string {
	missing := make([]string, 0)
	for _, key := range keys {
		if _, ok := numericPropertyOK(properties, key); !ok {
			missing = append(missing, key)
		}
	}
	return missing
}

func stringProperty(properties map[string]any, key string) string {
	if len(properties) == 0 {
		return ""
	}
	value, ok := properties[key]
	if !ok {
		return ""
	}
	if typed, ok := value.(string); ok {
		return strings.TrimSpace(typed)
	}
	return ""
}

func dynamicPathSegment(segment string) bool {
	if segment == "" {
		return false
	}
	if strings.HasPrefix(segment, ":") {
		return false
	}
	if len(segment) >= 12 {
		return true
	}
	digitCount := 0
	for _, r := range segment {
		if r >= '0' && r <= '9' {
			digitCount++
		}
	}
	return digitCount >= 4
}

func intPtr(value int) *int {
	return &value
}

func normalizeIntPtr(value *int) *int {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func intPtrString(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}
