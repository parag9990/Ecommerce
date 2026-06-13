package repository

import (
	"errors"
	"net/http"
	"testing"

	"github.com/typesense/typesense-go/v2/typesense"
)

func TestIsTypesenseStatus(t *testing.T) {
	err := &typesense.HTTPError{Status: http.StatusNotFound, Body: []byte(`{"message":"Not Found"}`)}
	if !isTypesenseStatus(err, http.StatusNotFound) {
		t.Fatal("expected 404 status match")
	}
	if isTypesenseStatus(err, http.StatusConflict) {
		t.Fatal("did not expect 409 status match")
	}
	if isTypesenseStatus(errors.New("status: 404"), http.StatusNotFound) {
		t.Fatal("plain string errors should not be treated as typed Typesense status")
	}
}
