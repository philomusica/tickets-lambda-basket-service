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
