package paymentHandler

type ErrPaymentFailed struct {
	Message string
}

func (e ErrPaymentFailed) Error() string {
	return e.Message
}
