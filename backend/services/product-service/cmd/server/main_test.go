package main

import (
	"net/http/httptest"
	"testing"
)

func TestAuthorizedInternalServiceFailsClosed(t *testing.T) {
	request := httptest.NewRequest("GET", "/internal/v1/products/search-export", nil)
	request.Header.Set("X-Service-Token", "service-token")

	if authorizedInternalService(request, "") {
		t.Fatal("empty configured token must fail closed")
	}
	if authorizedInternalService(request, "different-token") {
		t.Fatal("mismatched token must be rejected")
	}
	if !authorizedInternalService(request, "service-token") {
		t.Fatal("matching token must be accepted")
	}
}
