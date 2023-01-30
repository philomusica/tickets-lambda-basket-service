package main

import (
	"fmt"
	"testing"
	"github.com/aws/aws-lambda-go/events"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler"
)

var (
	paymentFailedStatusCode     int    = 404
	paymentFailedResponse       string = "Payment Failed. Please try again later"
	paymentSuccessfulStatusCode int    = 200
	paymentSuccessfulResponse   string = "payment successful"
	concertInPastErrMesage string = "Error concert x in the past, tickets are no longer avaiable"
)

func TestCalculateBalance(t *testing.T) {
	const (
		fullPriceCost    float32 = 11.0
		concessionCost   float32 = 9.0
		numOfFullPrice   uint8   = 2
		numOfConcessions uint8   = 2
	)
	result := calculateBalance(numOfFullPrice, fullPriceCost, numOfConcessions, concessionCost)
	var expectedResult float32 = 40
	if result != expectedResult {
		t.Errorf("Expected %.2f, got %.2f", expectedResult, result)
	}
}

func TestParseRequestBodySuccess(t *testing.T) {
	request := `
		{
			"numOfFullPrice": 1,
			"numOfConcessions": 2,
			"concertId": "ABC"
		}
	`
	var pr paymentRequest
	err := parseRequestBody(request, &pr)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}

func TestParseRequestBodyNoBody(t *testing.T) {
	request := ""
	var pr paymentRequest
	err := parseRequestBody(request, &pr)
	errMessage, ok := err.(ErrInvalidRequestBody)
	if !ok {
		t.Errorf("Expected err: '%s', got '%s'", err.(ErrInvalidRequestBody), errMessage)
	}
}

func TestParseRequestBodyInvalidFullPrice(t *testing.T) {
	request := `
		{
			"numOfFullPrice": "bob",
			"numOfConcessions": 2,
			"concertId": "ABC"
		}
	`
	var pr paymentRequest
	err := parseRequestBody(request, &pr)
	errMessage, ok := err.(ErrInvalidRequestBody)
	if !ok {
		t.Errorf("Expected err: '%s', got '%s'", err.(ErrInvalidRequestBody), errMessage)
	}
}

func TestParseRequestBodyNoConcessionPrice(t *testing.T) {
	request := `
		{
			"numOfFullPrice": 2,
			"concertId": "ABC"
		}
	`
	var pr paymentRequest
	err := parseRequestBody(request, &pr)
	errMessage, ok := err.(ErrInvalidRequestBody)
	if !ok {
		t.Errorf("Expected err: '%s', got '%s'", err.(ErrInvalidRequestBody), errMessage)
	}
}

func TestParseRequestBodyInvalidConcertId(t *testing.T) {
	request := `
		{
			"numOfFullPrice": 2,
			"numOfConcessions": 2,
			"concertId": 2
		}
	`
	var pr paymentRequest
	err := parseRequestBody(request, &pr)
	errMessage, ok := err.(ErrInvalidRequestBody)
	if !ok {
		t.Errorf("Expected err: '%s', got '%s'", err.(ErrInvalidRequestBody), errMessage)
	}
}

func TestHandlerInvalidRequest(t *testing.T) {
	request := events.APIGatewayProxyRequest{
		Body: `{
			"numOfConcessions": 2,
			"concertId": "ABC"	
		}`,
	}

	response := Handler(request)
	expectedStatusCode := 400
	expectedResponseBody := "Invalid request"
	if response.StatusCode != expectedStatusCode || response.Body != expectedResponseBody {
		t.Errorf("Expected statusCode %d and Body %s, got %d and %s", expectedStatusCode, expectedResponseBody, response.StatusCode, response.Body)
	}
}

type mockDDBHandlerConcertInPast struct {
	databaseHandler.DatabaseHandler
}

func (m mockDDBHandlerConcertInPast) GetConcertFromDatabase(concertID string) (concert *databaseHandler.Concert, err error) {
	err = databaseHandler.ErrConcertInPast{Message: "Error concert x in the past, tickets are no longer avaiable"}
	return
}

func TestProcessPaymentConcertInPast(t *testing.T) {

	request := events.APIGatewayProxyRequest{
		Body: `{
			"numOfFullPrice": 2,
			"numOfConcessions": 2,
			"concertId": "ABC"
		}`,
	}

	mockHandler := mockDDBHandlerConcertInPast{}
	response := processPayment(request, mockHandler)

	expectedStatusCode := 400
	if response.StatusCode != expectedStatusCode || response.Body != concertInPastErrMesage {
		t.Errorf("Expected statusCode %d and Body %s, got %d and %s", expectedStatusCode, concertInPastErrMesage, response.StatusCode, response.Body)
	}
}

type mockDDBHandlerInsufficientTickets struct {
	databaseHandler.DatabaseHandler
}

func (m mockDDBHandlerInsufficientTickets) GetConcertFromDatabase(concertID string) (concert *databaseHandler.Concert, err error) {
	concert = &databaseHandler.Concert{
		AvailableTickets: 9,
		Description: "summer concert",
	}
	return
}

func TestProcessPaymentInsufficientTicketsAvailable(t *testing.T) {
	request := events.APIGatewayProxyRequest{
		Body: `{
			"numOfFullPrice": 8,
			"numOfConcessions": 2,
			"concertId": "ABC"
		}`,
	}

	mockHandler := mockDDBHandlerInsufficientTickets{}
	response := processPayment(request, mockHandler)

	expectedStatusCode := 403
	expectedResponseBody := fmt.Sprintf("Insufficient tickets available for %s\n", "summer concert")
	if response.StatusCode != expectedStatusCode || response.Body != expectedResponseBody {
		t.Errorf("Expected statusCode %d and Body %s, got %d and %s", expectedStatusCode, expectedResponseBody, response.StatusCode, response.Body)
	}
}
