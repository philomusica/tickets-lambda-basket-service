package paymentHandler

type PaymentHandler interface {
	Process(balance float32, reference string) (clientSecret string, err error)
}

type OrderLine struct {
	ConcertID        string `json:"concertID"`
	NumOfFullPrice   *uint8 `json:"numOfFullPrice"`
	NumOfConcessions *uint8 `json:"numOfConcessions"`
}

type Order struct {
	ConcertID        string `json:"concertID"`
	OrderReference   string `json:"orderReference"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	Email            string `json:"email"`
	NumOfFullPrice   uint8  `json:"numOfFullPrice"`
	NumOfConcessions uint8  `json:"numOfConcessions"`
	OrderStatus      string `json:"orderStatus"`
}

// PaymentRequest is a struct representing the json object passed to the lambda containing ticket and payment details
type PaymentRequest struct {
	OrderLines []OrderLine `json:"orderLines"`
	FirstName  string      `json:"firstName"`
	LastName   string      `json:"lastName"`
	Email      string      `json:"email"`
}
