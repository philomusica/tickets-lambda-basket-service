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
	var expectedResult float32 = (float32(numOfFullPrice)*fullPriceCost + float32(numOfConcessions)*concessionCost)
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
