package device

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"github.com/oschwald/geoip2-golang"
)

type DisabledGeoResolver struct{}

func (DisabledGeoResolver) Resolve(ctx context.Context, clientIP string) domain.Geo {
	return domain.Geo{Source: domain.GeoSourceDisabled}
}

type MaxMindGeoResolver struct {
	reader *geoip2.Reader
}

func NewMaxMindGeoResolver(dbPath string) (*MaxMindGeoResolver, error) {
	dbPath = strings.TrimSpace(dbPath)
	if dbPath == "" {
		return nil, errors.New("geoip database path is required")
	}
	reader, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open geoip database: %w", err)
	}
	return &MaxMindGeoResolver{reader: reader}, nil
}

func (r *MaxMindGeoResolver) Resolve(ctx context.Context, clientIP string) domain.Geo {
	if r == nil || r.reader == nil {
		return domain.Geo{Source: domain.GeoSourceDisabled}
	}
	if err := ctx.Err(); err != nil {
		return domain.Geo{Source: domain.GeoSourceUnknown}
	}
	ip := net.ParseIP(strings.TrimSpace(clientIP))
	if ip == nil || privateOrLoopback(ip) {
		return domain.Geo{Source: domain.GeoSourceUnknown}
	}
	city, err := r.reader.City(ip)
	if err != nil {
		return r.resolveCountry(ip)
	}
	geo := domain.Geo{
		Country:  stringPtr(city.Country.IsoCode),
		Timezone: stringPtr(city.Location.TimeZone),
		Source:   domain.GeoSourceGeoIP,
	}
	if len(city.Subdivisions) > 0 {
		geo.Region = stringPtr(city.Subdivisions[0].IsoCode)
	}
	if name := cityName(city.City.Names); name != "" {
		geo.City = &name
	}
	return geo.Normalize()
}

func (r *MaxMindGeoResolver) resolveCountry(ip net.IP) domain.Geo {
	country, err := r.reader.Country(ip)
	if err != nil {
		return domain.Geo{Source: domain.GeoSourceUnknown}
	}
	return domain.Geo{
		Country: stringPtr(country.Country.IsoCode),
		Source:  domain.GeoSourceGeoIP,
	}.Normalize()
}

func (r *MaxMindGeoResolver) Close() error {
	if r == nil || r.reader == nil {
		return nil
	}
	return r.reader.Close()
}

func privateOrLoopback(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsMulticast()
}

func cityName(names map[string]string) string {
	if len(names) == 0 {
		return ""
	}
	if name := strings.TrimSpace(names["en"]); name != "" {
		return name
	}
	for _, name := range names {
		if trimmed := strings.TrimSpace(name); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func stringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
