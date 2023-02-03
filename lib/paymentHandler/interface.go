package paymentHandler

type PaymentHandler interface {
	Process(payReq PaymentRequest, balance float32) (err error)
}

type OrderLine struct {
	ConcertId        string `json:"concertId"`
	NumOfFullPrice   *uint8 `json:"numOfFullPrice"`
	NumOfConcessions *uint8 `json:"numOfConcessions"`
}

type Order struct {
	Reference string `jsong:"reference"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Email string `json:"email"`
	OL OrderLine `json:"ol"`
}

// PaymentRequest is a struct representing the json object passed to the lambda containing ticket and payment details
type PaymentRequest struct {
	OrderLines []OrderLine `json:"orderLines"`
	EmailAddress string `json:"emailAddress"`
}
