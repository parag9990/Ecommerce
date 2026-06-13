package domain

import "strings"

type EnrichedDeviceContext struct {
	UserAgent             string    `json:"user_agent"`
	UserAgentHash         *string   `json:"user_agent_hash,omitempty"`
	IPHash                string    `json:"ip_hash"`
	IPVersion             IPVersion `json:"ip_version,omitempty"`
	DeviceFingerprintHash *string   `json:"device_fingerprint_hash,omitempty"`
	Device                Device    `json:"device"`
	Client                Client    `json:"client"`
	Geo                   Geo       `json:"geo"`
}

func (c EnrichedDeviceContext) Normalize() EnrichedDeviceContext {
	out := c
	out.UserAgent = strings.TrimSpace(out.UserAgent)
	out.UserAgentHash = trimStringPtr(out.UserAgentHash)
	out.IPHash = strings.TrimSpace(out.IPHash)
	out.IPVersion = IPVersion(strings.TrimSpace(string(out.IPVersion)))
	out.DeviceFingerprintHash = trimStringPtr(out.DeviceFingerprintHash)
	out.Device = out.Device.Normalize()
	if out.Device.Type == "" {
		out.Device.Type = DeviceTypeUnknown
	}
	out.Client = out.Client.Normalize()
	out.Geo = out.Geo.Normalize()
	return out
}
