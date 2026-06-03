package validation

import "testing"

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{name: "india local", raw: "99999 99999", want: "+919999999999", ok: true},
		{name: "e164", raw: "+919999999999", want: "+919999999999", ok: true},
		{name: "bad characters", raw: "abc", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := NormalizePhone(tt.raw, "IN")
			if ok != tt.ok || got != tt.want {
				t.Fatalf("NormalizePhone() = %q, %t; want %q, %t", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestIsValidGSTIN(t *testing.T) {
	tests := []struct {
		name string
		gst  string
		want bool
	}{
		{name: "valid", gst: "22AAAAA0000A1Z5", want: true},
		{name: "lowercase normalized", gst: "22aaaaa0000a1z5", want: true},
		{name: "bad state", gst: "99AAAAA0000A1Z5", want: false},
		{name: "too short", gst: "22AAAAA0000A1Z", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidGSTIN(tt.gst); got != tt.want {
				t.Fatalf("IsValidGSTIN() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestPostalCodeByCountry(t *testing.T) {
	if !IsValidPostalCode("IN", "560001") {
		t.Fatal("expected valid India PIN")
	}
	if IsValidPostalCode("IN", "012345") {
		t.Fatal("expected invalid India PIN")
	}
	if !IsValidPostalCode("US", "94105-1234") {
		t.Fatal("expected valid US ZIP+4")
	}
}
