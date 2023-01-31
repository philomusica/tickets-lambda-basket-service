package stripePaymentHandler

import (
	"testing"
	"github.com/philomusica/tickets-lambda-process-payment/lib/paymentHandler"
)

func TestStripePaymentHandlerFails(t *testing.T) {
	stripeHandler := New()

	payReq := paymentHandler.PaymentRequest{}
	var balance float32 = 40.0
	err := stripeHandler.Process(payReq, balance)
	expectedErr, ok := err.(paymentHandler.ErrPaymentFailed)
	
	if !ok {
		t.Errorf("Expected err %T, got %T\n", expectedErr, err)
	}
}

