package http

import "net/http"

type RouteRegistrar interface {
	Register(mux *http.ServeMux)
}

func NewServeMux(handler *RBACHandler, routeRegistrars ...RouteRegistrar) *http.ServeMux {
	mux := http.NewServeMux()
	handler.Register(mux)
	for _, registrar := range routeRegistrars {
		if registrar != nil {
			registrar.Register(mux)
		}
	}
	return mux
}
