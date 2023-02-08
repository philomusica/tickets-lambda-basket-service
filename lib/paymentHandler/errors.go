package paymentHandler

// ErrPaymentFailed is a custom error to signify the payment failed
type ErrPaymentFailed struct {
	Message string
}

func (e ErrPaymentFailed) Error() string {
	return e.Message
}

// ErrOrderDoesNotExist is a custom error to signify the order does not exist in the database
type ErrOrderDoesNotExist struct {
	Message string
}

func (e ErrOrderDoesNotExist) Error() string {
	return e.Message
}
