package httptransport

type ResponseEnvelope struct {
	Data      any       `json:"data"`
	RequestID string    `json:"request_id"`
	Error     *ErrorDTO `json:"error"`
}

type ErrorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type ErrorDetailDTO struct {
	Field   string `json:"field,omitempty"`
	Reason  string `json:"reason"`
	Message string `json:"message,omitempty"`
}

type HealthDTO struct {
	Status       string            `json:"status"`
	Service      string            `json:"service"`
	Routes       int               `json:"routes,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}
