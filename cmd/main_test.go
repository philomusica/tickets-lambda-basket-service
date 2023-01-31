package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler"
	"github.com/philomusica/tickets-lambda-process-payment/lib/paymentHandler"
	"testing"
)

var (
	concertInPastErrMesage      string = "Error concert x in the past, tickets are no longer avaiable"
)

func TestParseRequestBodySuccess(t *testing.T) {
	request := `
		{
			"orderLines": [
				{
					"numOfFullPrice": 1,
					"numOfConcessions": 2,
					"concertId": "ABC"
				}
			]
		}
	`
	var pr paymentHandler.PaymentRequest
	err := parseRequestBody(request, &pr)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}

func TestParseRequestBodyNoBody(t *testing.T) {
	request := ""
	var pr paymentHandler.PaymentRequest
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
	var pr paymentHandler.PaymentRequest
	err := parseRequestBody(request, &pr)
	errMessage, ok := err.(ErrInvalidRequestBody)
	if !ok {
		t.Errorf("Expected err: '%s', got '%s'", err.(ErrInvalidRequestBody), errMessage)
	}
}

func TestParseRequestBodyNoConcessionPrice(t *testing.T) {
	request := `
		"orderLines": [
			{
				"numOfFullPrice": 2,
				"concertId": "ABC"
			}
		]
	`
	var pr paymentHandler.PaymentRequest
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
	var pr paymentHandler.PaymentRequest
	err := parseRequestBody(request, &pr)
	errMessage, ok := err.(ErrInvalidRequestBody)
	if !ok {
		t.Errorf("Expected err: '%s', got '%s'", err.(ErrInvalidRequestBody), errMessage)
	}
}

func TestHandlerInvalidRequest(t *testing.T) {
	request := events.APIGatewayProxyRequest{
		Body: `{
			"orderLines": [
				{
					"numOfConcessions": 2,
					"concertId": "ABC"	
				}
			]
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

type mockStripeHandlerEmpty struct {}

func (m mockStripeHandlerEmpty) Process(payReq paymentHandler.PaymentRequest, balance float32) (err error) {
	return
}

func TestProcessPaymentConcertInPast(t *testing.T) {

	request := events.APIGatewayProxyRequest{
		Body: `{
			"orderLines": [
				{
					"numOfFullPrice": 2,
					"numOfConcessions": 2,
					"concertId": "ABC"
				}
			]
		}`,
	}

	mockDyanmoHanlder := mockDDBHandlerConcertInPast{}
	mockStripeHandler := mockStripeHandlerEmpty{}
	response := processPayment(request, mockDyanmoHanlder, mockStripeHandler)

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
		Description:      "summer concert",
	}
	return
}

func TestProcessPaymentInsufficientTicketsAvailable(t *testing.T) {
	request := events.APIGatewayProxyRequest{
		Body: `{
			"orderLines": [
				{
					"numOfFullPrice": 8,
					"numOfConcessions": 2,
					"concertId": "ABC"
				}
			]
		}`,
	}

	mockDynamoHandler := mockDDBHandlerInsufficientTickets{}
	mockStripeHandler := mockStripeHandlerEmpty{}
	response := processPayment(request, mockDynamoHandler, mockStripeHandler)

	expectedStatusCode := 403
	expectedResponseBody := fmt.Sprintf("Insufficient tickets available for %s\n", "summer concert")
	if response.StatusCode != expectedStatusCode || response.Body != expectedResponseBody {
		t.Errorf("Expected statusCode %d and Body %s, got %d and %s", expectedStatusCode, expectedResponseBody, response.StatusCode, response.Body)
	}
}
