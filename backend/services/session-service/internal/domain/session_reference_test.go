package domain

import "testing"

func TestSessionReferenceCodecRoundTripAndTamperRejection(t *testing.T) {
	codec, err := NewSessionReferenceCodec("a-long-test-privacy-secret")
	if err != nil {
		t.Fatalf("new codec: %v", err)
	}
	reference, err := codec.Protect("sess_123")
	if err != nil {
		t.Fatalf("protect: %v", err)
	}
	if !IsSessionReference(reference) || reference == "sess_123" {
		t.Fatalf("unexpected reference %q", reference)
	}
	resolved, err := codec.Unprotect(reference)
	if err != nil || resolved != "sess_123" {
		t.Fatalf("round trip = %q, %v", resolved, err)
	}
	if _, err := codec.Unprotect(reference + "x"); err == nil {
		t.Fatal("tampered reference was accepted")
	}
}
