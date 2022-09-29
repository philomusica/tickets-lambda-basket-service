package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type ErrInvalidRequestBody struct {
	Message string
}

func (e ErrInvalidRequestBody) Error() string {
	return e.Message
}

// PaymentRequest is a struct representing the json object passed to the lambda containing ticket and payment details
type PaymentRequest struct {
	NumOfFullPrice   *uint8 `json:"numOfFullPrice"`
	NumOfConcessions *uint8 `json:"numOfConcessions"`
}

// CalculateBalance calculates the total due for the num of tickets ordered, taking into account different ticket costs
func CalculateBalance(numOfFullPrice uint8, fullPriceCost float32, numOfConcessions uint8, concessionCost float32) (total float32) {
	total = float32(numOfConcessions)*fullPriceCost + float32(numOfConcessions)*concessionCost
	return
}

// ParseRequestBody takes the request body as string and unmarshals it into the PaymentRequest struct
func ParseRequestBody(request string, payReq *PaymentRequest) (err error) {
	br := []byte(request)
	err = json.Unmarshal(br, payReq)
	if payReq.NumOfFullPrice == nil || payReq.NumOfConcessions == nil {
		err = ErrInvalidRequestBody{Message: "no value present for num of full price or num of concession tickets"}
	}
	return
}

// Handler function is the entry point for the lambda function
func Handler(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	fullPriceCostStr  := os.Getenv("FULL_PRICE_COST")
	concessionCostStr := os.Getenv("CONCESSION_COST")
	response := events.APIGatewayProxyResponse{
		StatusCode: 404,
		Body:       "Payment Failed. Please try again later",
	}

	var payReq PaymentRequest
	err := ParseRequestBody(request.Body, &payReq)
	if err != nil {
		fmt.Println(err)
		return response
	}

	fullPrice, err := strconv.ParseFloat(fullPriceCostStr, 32)
	if err != nil {
		fmt.Printf("Unable to parse %s as float\n", fullPriceCostStr)
		return response
	}

	concessionPrice, err := strconv.ParseFloat(concessionCostStr, 32)
	if err != nil {
		fmt.Printf("Unable to parse %s as float\n", fullPriceCostStr)
		return response
	}

	result := CalculateBalance(*payReq.NumOfFullPrice, float32(fullPrice), *payReq.NumOfConcessions, float32(concessionPrice))
	fmt.Printf("Result is + %f\n", result)

	response.StatusCode = 200
	response.Body = "payment successful"
	return response
}
func main() {
	lambda.Start(Handler)
}
