package emailHandler

import (
	"github.com/philomusica/tickets-lambda-process-payment/lib/paymentHandler"
)

type EmailHandler interface {
	SendEmail(order paymentHandler.Order, attachment []byte) (err error)
}


