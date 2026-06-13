package httptransport

import "net/http"

func NewRouter(handler *Handler) http.Handler {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return mux
}
