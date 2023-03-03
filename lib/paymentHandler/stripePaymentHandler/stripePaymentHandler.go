package stripePaymentHandler

import (
	"fmt"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
)

// ===============================================================================================================================
// TYPE DEFINITIONS
// ===============================================================================================================================

type StripePaymentHandler struct {
	stripeSecret string
}

// ===============================================================================================================================
// END TYPE DEFINITIONS
// ===============================================================================================================================

// ===============================================================================================================================
// PUBLIC FUNCTIONS
// ===============================================================================================================================

func New(stripeSecret string) (sph *StripePaymentHandler) {
	return &StripePaymentHandler{
		stripeSecret,
	}
}

func (s StripePaymentHandler) Process(balance float32, reference string) (clientSecret string, err error) {
	stripe.Key = s.stripeSecret
	params := &stripe.PaymentIntentParams{

		Amount:   stripe.Int64(int64(balance * 100)),
		Currency: stripe.String(string(stripe.CurrencyGBP)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}
	params.AddMetadata("order_reference", reference)

	intent, err := paymentintent.New(params)
	if err != nil {
		fmt.Println(err)
		return
	}
	clientSecret = intent.ClientSecret
	return
}

// ===============================================================================================================================
// END PUBLIC FUNCTIONS
// ===============================================================================================================================
