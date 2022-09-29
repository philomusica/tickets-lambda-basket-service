package main

import (
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
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
			"numOfConcessions": 2
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

func TestHandlerPaymentRequestValid(t *testing.T) {

	os.Setenv("FULL_PRICE_COST", "15.00")
	os.Setenv("CONCESSION_COST", "10.00")

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

func TestHandlerCannotParseFullPriceCostEnvVar(t *testing.T) {

	os.Setenv("FULL_PRICE_COST", "blah")
	os.Setenv("CONCESSION_COST", "10.00")

	request := events.APIGatewayProxyRequest{
		Body: `{
			"numOfFullPrice": 2,
			"numOfConcessions": 2
		}`,
	}

	response := Handler(request)
	if response.StatusCode != 404 || response.Body != "Payment Failed. Please try again later" {
		t.Errorf("Expected StatusCode 200 and response of \"payment successful\", got %d and %s", response.StatusCode, response.Body)
	}
}

func TestHandlerCannotParseConcessionEnvVar(t *testing.T) {

	os.Setenv("FULL_PRICE_COST", "15.00")
	os.Setenv("CONCESSION_COST", "blah")

	request := events.APIGatewayProxyRequest{
		Body: `{
			"numOfFullPrice": 2,
			"numOfConcessions": 2
		}`,
	}

	response := Handler(request)
	if response.StatusCode != 404 || response.Body != "Payment Failed. Please try again later" {
		t.Errorf("Expected StatusCode 200 and response of \"payment successful\", got %d and %s", response.StatusCode, response.Body)
	}
}
func TestHandlerNoNumOfFullPriceInJson(t *testing.T) {

	os.Setenv("FULL_PRICE_COST", "blah")
	os.Setenv("CONCESSION_COST", "bob")

	request := events.APIGatewayProxyRequest{
		Body: `{
			"numOfConcessions": 2
		}`,
	}

	response := Handler(request)
	if response.StatusCode != 404 || response.Body != "Payment Failed. Please try again later" {
		t.Errorf("Expected StatusCode 200 and response of \"payment successful\", got %d and %s", response.StatusCode, response.Body)
	}
}

func TestHandlerNoNumOfConcessionInJson(t *testing.T) {

	os.Setenv("FULL_PRICE_COST", "blah")
	os.Setenv("CONCESSION_COST", "bob")

	request := events.APIGatewayProxyRequest{
		Body: `{
			"numOfFullPrice": 2
		}`,
	}

	response := Handler(request)
	if response.StatusCode != 404 || response.Body != "Payment Failed. Please try again later" {
		t.Errorf("Expected StatusCode 200 and response of \"payment successful\", got %d and %s", response.StatusCode, response.Body)
	}
}