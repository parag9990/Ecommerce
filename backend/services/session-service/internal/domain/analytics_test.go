package domain

import "testing"

func TestCalculateFunnelRates(t *testing.T) {
	steps := []FunnelStep{
		{Name: "product_view", EventType: EventProductView, UniqueSessions: 100},
		{Name: "add_to_cart", EventType: EventAddToCart, UniqueSessions: 25},
		{Name: "payment_result", EventType: EventPaymentResult, UniqueSessions: 5},
	}

	got := CalculateFunnelRates(steps)

	if got[0].ConversionFromPrevious != 100 || got[0].DropoffFromPrevious != 0 {
		t.Fatalf("unexpected first step rates: %+v", got[0])
	}
	if got[1].ConversionFromPrevious != 25 || got[1].DropoffFromPrevious != 75 {
		t.Fatalf("unexpected second step rates: %+v", got[1])
	}
	if got[2].ConversionFromPrevious != 20 || got[2].DropoffFromPrevious != 80 {
		t.Fatalf("unexpected third step rates: %+v", got[2])
	}
	if OverallFunnelConversion(got) != 5 {
		t.Fatalf("unexpected overall conversion: %v", OverallFunnelConversion(got))
	}
}

func TestCalculateFunnelRatesHandlesZeroPreviousStep(t *testing.T) {
	steps := []FunnelStep{
		{Name: "product_view", EventType: EventProductView, UniqueSessions: 0},
		{Name: "add_to_cart", EventType: EventAddToCart, UniqueSessions: 25},
	}

	got := CalculateFunnelRates(steps)

	if got[1].ConversionFromPrevious != 0 || got[1].DropoffFromPrevious != 100 {
		t.Fatalf("unexpected zero-division rates: %+v", got[1])
	}
}

func TestValidateFunnelStepTypesRejectsUnsupportedStep(t *testing.T) {
	err := ValidateFunnelStepTypes([]EventType{EventProductView, EventType("password_submit")}, 2, 8)
	if err == nil {
		t.Fatal("expected unsupported funnel step error")
	}
}
