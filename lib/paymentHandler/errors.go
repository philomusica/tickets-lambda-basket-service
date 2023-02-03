package paymentHandler

type ErrPaymentFailed struct {
	Message string
}

func (e ErrPaymentFailed) Error() string {
	return e.Message
}

type ErrOrderDoesNotExist struct {
	Message string
}

func (e ErrOrderDoesNotExist) Error() string {
	return e.Message
}
