package httptransport

import "net/http"

func NewRouter(handler *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return mux
}
