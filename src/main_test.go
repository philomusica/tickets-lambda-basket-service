package main

import (
	"testing"
)

func TestCalculateBalance(t *testing.T) {
	const (
		fullPriceCost    float32 = 11.0
		concessionCost   float32 = 9.0
		numOfFullPrice   uint8   = 2
		numOfConcessions uint8   = 2
	)
	result := CalculateBalance(numOfFullPrice, fullPriceCost, numOfConcessions, concessionCost)
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
	var pr PaymentRequest
	err := ParseRequestBody(request, &pr)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}

func TestParseRequestBodyNoBody(t *testing.T) {
	request := ""
	var pr PaymentRequest
	err := ParseRequestBody(request, &pr)
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
	var pr PaymentRequest
	err := ParseRequestBody(request, &pr)
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
	var pr PaymentRequest
	err := ParseRequestBody(request, &pr)
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
	var pr PaymentRequest
	err := ParseRequestBody(request, &pr)
	errMessage, ok := err.(ErrInvalidRequestBody)
	if !ok {
		t.Errorf("Expected err: '%s', got '%s'", err.(ErrInvalidRequestBody), errMessage)
	}
}

/*
func TestHandlerPaymentRequestValid(t *testing.T) {

	request := events.APIGatewayProxyRequest{
		Body: `{
			"numOfFullPrice": 2,
			"numOfConcessions": 2
		}`,
	}

	response := Handler(request)
	if response.StatusCode != 200 || response.Body != "payment successful" {
		t.Errorf("Expected StatusCode 200 and response of \"payment successful\", got %d and %s", response.StatusCode, response.Body)
	}
}
*/
