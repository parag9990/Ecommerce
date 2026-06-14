package domain

// TemplateKey identifies a business-controlled template family.
type TemplateKey string

const (
	TemplateOTPVerification     TemplateKey = "otp_verification"
	TemplateOrderStatusUpdate   TemplateKey = "order_status_update"
	TemplatePaymentStatusUpdate TemplateKey = "payment_status_update"
	TemplateWelcomeUser         TemplateKey = "welcome_user"
	TemplateSellerApproved      TemplateKey = "seller_approved"
	TemplateAddressUpdated      TemplateKey = "address_updated_security_notice"
	TemplatePromotionalOffer    TemplateKey = "promotional_offer"
)

// RenderRequest contains the approved input boundary for template rendering.
type RenderRequest struct {
	TemplateKey TemplateKey
	Channel     Channel
	Variables   map[string]string
}

// RenderedMessage is provider-independent rendered content.
type RenderedMessage struct {
	TemplateID      string
	TemplateKey     TemplateKey
	TemplateVersion int
	Channel         Channel
	Subject         string
	Body            string
}
