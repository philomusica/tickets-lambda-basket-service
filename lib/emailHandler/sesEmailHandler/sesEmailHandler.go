package sesEmailHandler

import (
	"github.com/aws/aws-sdk-go/service/sesv2/sesv2iface"
	"github.com/philomusica/tickets-lambda-process-payment/lib/paymentHandler"
)

// ===============================================================================================================================
// TYPE DEFINITIONS
// ===============================================================================================================================

type SESEmailHandler struct {
	svc sesv2iface.SESV2API
}

// ===============================================================================================================================
// END TYPE DEFINITIONS
// ===============================================================================================================================

// ===============================================================================================================================
// PUBLIC FUNCTIONS
// ===============================================================================================================================

// New takes an SES V2 interface and returns a newly created SESEmailHandler struct
func New(svc sesv2iface.SESV2API) SESEmailHandler {
	return SESEmailHandler{
		svc,
	}
}

// SendEmail takes an order struct and attachment (in bytes) and sends an email to the customer, using the AWS SES v2 API. Returns an error if fails, or nil if successful
func (s SESEmailHandler) SendEmail(order paymentHandler.Order, attachment []byte) (err error) {
	return
}
// ===============================================================================================================================
// END PUBLIC FUNCTIONS
// ===============================================================================================================================
