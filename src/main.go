package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler/ddbHandler"
)

type ErrInvalidRequestBody struct {
	Message string
}

func (e ErrInvalidRequestBody) Error() string {
	return e.Message
}

// paymentRequest is a struct representing the json object passed to the lambda containing ticket and payment details
type paymentRequest struct {
	ConcertId        string `json:"concertId"`
	NumOfFullPrice   *uint8 `json:"numOfFullPrice"`
	NumOfConcessions *uint8 `json:"numOfConcessions"`
}

// calculateBalance calculates the total due for the num of tickets ordered, taking into account different ticket costs
func calculateBalance(numOfFullPrice uint8, fullPriceCost float32, numOfConcessions uint8, concessionCost float32) (total float32) {
	total = float32(numOfConcessions)*fullPriceCost + float32(numOfConcessions)*concessionCost
	return
}

// parseRequestBody takes the request body as string and unmarshals it into the PaymentRequest struct
func parseRequestBody(request string, payReq *paymentRequest) (err error) {
	br := []byte(request)
	err = json.Unmarshal(br, payReq)
	if err != nil || payReq.NumOfFullPrice == nil || payReq.NumOfConcessions == nil || payReq.ConcertId == "" {
		err = ErrInvalidRequestBody{Message: "no value present for num of full price or num of concession tickets"}
	}
	return
}

func processPayment(request events.APIGatewayProxyRequest, handler databaseHandler.DatabaseHandler) (err error) {
	var payReq paymentRequest
	err = parseRequestBody(request.Body, &payReq)
	if err != nil {
		return
	}

	var concert *databaseHandler.Concert
	concert, err = handler.GetConcertFromDatabase(payReq.ConcertId)
	if err != nil {
		return
	}

	_ = calculateBalance(*payReq.NumOfFullPrice, concert.FullPrice, *payReq.NumOfConcessions, concert.ConcessionPrice)
	return
}

// Handler function is the entry point for the lambda function
func Handler(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	response := events.APIGatewayProxyResponse{
		StatusCode: 404,
		Body:       "Payment Failed. Please try again later",
	}

	sess := session.New()
	svc := dynamodb.New(sess)
	handler := ddbHandler.New(svc)

	err := processPayment(request, handler)
	if err != nil {
		fmt.Println(err)
		return response
	}

	response.StatusCode = 200
	response.Body = "payment successful"
	return response
}
func main() {
	lambda.Start(Handler)
}
