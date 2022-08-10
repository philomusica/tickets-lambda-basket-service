package main

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	fullPriceCostStr  = os.Getenv("FULL_PRICE_COST")
	concessionCostStr = os.Getenv("CONCESSION_COST")
)

type PaymentRequest struct {
	NumOfFullPrice   uint8
	numOfConcessions uint8
}

func CalculateBalance(numOfFullPrice uint8, fullPriceCost float32, numOfConcessions uint8, concessionCost float32) (total float32) {
	total = float32(numOfConcessions)*fullPriceCost + float32(numOfConcessions)*concessionCost
	return total
}

func ParseRequestBody(request string, payReq *PaymentRequest) (err error) {
	br := []byte(request)
	err = json.Unmarshal(br, payReq)
	return
}

func Handler(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	response := events.APIGatewayProxyResponse{
		StatusCode: 404,
		Body:       "Payment Failed. Please try again later",
	}

	var payReq PaymentRequest
	ParseRequestBody(request.Body, &payReq)

	return response
}
func main() {
	lambda.Start(Handler)
}
