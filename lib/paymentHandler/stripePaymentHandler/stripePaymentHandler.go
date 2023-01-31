package stripePaymentHandler

import (
	"github.com/philomusica/tickets-lambda-process-payment/lib/paymentHandler"
)

type StripePaymentHandler struct {}

func New() (sph *StripePaymentHandler) {
	return &StripePaymentHandler{}
}

func (s StripePaymentHandler) Process(payReq paymentHandler.PaymentRequest, balance float32) (err error) {
	err = paymentHandler.ErrPaymentFailed{}
	return
}
