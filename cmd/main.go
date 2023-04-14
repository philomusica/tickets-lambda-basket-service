package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/philomusica/tickets-lambda-basket-service/lib/paymentHandler"
	"github.com/philomusica/tickets-lambda-basket-service/lib/paymentHandler/stripePaymentHandler"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler/ddbHandler"
)

// ===============================================================================================================================
// TYPE DEFINITIONS
// ===============================================================================================================================

// ErrInvalidRequestBody is a custom error to signify the payment request JSON is invalid
type ErrInvalidRequestBody struct {
	Message string
}

func (e ErrInvalidRequestBody) Error() string {
	return e.Message
}

// ErrInsufficientAvailableTickets is a custom error to signify there aren't sufficient tickets available for a given concert
type ErrInsufficientAvailableTickets struct {
	Message string
}

func (e ErrInsufficientAvailableTickets) Error() string {
	return e.Message
}

// ===============================================================================================================================
// END TYPE DEFINITIONS
// ===============================================================================================================================

// ===============================================================================================================================
// PRIVATE FUNCTIONS
// ===============================================================================================================================

// parseRequestBody takes the request body as string and unmarshals it into the PaymentRequest struct
func parseRequestBody(request string, payReq *paymentHandler.PaymentRequest) (err error) {
	br := []byte(request)
	err = json.Unmarshal(br, payReq)
	if err != nil {
		err = ErrInvalidRequestBody{Message: err.Error()}
		return
	}
	if len(payReq.OrderLines) == 0 {
		err = ErrInvalidRequestBody{Message: "No orders made"}
		return
	}

	for _, ol := range payReq.OrderLines {
		if ol.NumOfFullPrice == nil || ol.NumOfConcessions == nil || ol.ConcertID == "" {
			err = ErrInvalidRequestBody{Message: "order line is missing requirement information"}
			return
		}
	}
	return
}

// processPayment is the main function, taking the AWS events.APIGatewayProxyRequest struct, a DatabaseHandler and PaymentHandler (both interfaces) and response an AWS events.APIGatewayProxyResponse struct
func processPayment(request events.APIGatewayProxyRequest, dbHandler databaseHandler.DatabaseHandler, payHandler paymentHandler.PaymentHandler) (response events.APIGatewayProxyResponse) {
	response.Headers = map[string]string{"Access-Control-Allow-Origin": "*"}
	var payReq paymentHandler.PaymentRequest
	err := parseRequestBody(request.Body, &payReq)
	if err != nil {
		fmt.Println(err)
		response.StatusCode = 400
		response.Body = "Invalid request"
		return
	}

	var balance float32 = 0.0

	concerts := make(map[string]databaseHandler.Concert)

	for _, ol := range payReq.OrderLines {
		var concert *databaseHandler.Concert
		concert, err = dbHandler.GetConcertFromTable(ol.ConcertID)
		if err != nil {
			fmt.Println(err)
			response.StatusCode = 400
			response.Body = err.Error()
			return
		}

		ticketTotal := uint16(*ol.NumOfFullPrice + *ol.NumOfConcessions)
		if concert.AvailableTickets < ticketTotal {
			err = ErrInsufficientAvailableTickets{Message: fmt.Sprintf("Insufficient tickets available for %s\n", concert.Title)}
			fmt.Println(err)
			response.StatusCode = 403
			response.Body = err.Error()
			return
		}

		balance += float32(*ol.NumOfFullPrice)*concert.FullPrice + float32(*ol.NumOfConcessions)*concert.ConcessionPrice
		concerts[ol.ConcertID] = *concert
	}

	reference := dbHandler.GenerateOrderReference(4)
	for _, ol := range payReq.OrderLines {
		// Set default error message
		errMessage := "Internal Server Error"
		response.StatusCode = 500
		response.Body = errMessage

		// Create Order struct
		order := paymentHandler.Order{
			Reference:        reference,
			ConcertID:        ol.ConcertID,
			FirstName:        payReq.FirstName,
			LastName:         payReq.LastName,
			Email:            payReq.Email,
			NumOfFullPrice:   *ol.NumOfFullPrice,
			NumOfConcessions: *ol.NumOfConcessions,
			Status:           "pending",
		}

		// Add order to orders table
		err = dbHandler.CreateOrderInTable(order)
		if err != nil {
			fmt.Printf("Unable to create order in Orders table: %s\n", err)
			return
		}
	}

	clientSecret, err := payHandler.Process(balance, reference)
	if err != nil {
		fmt.Println(err)
		response.StatusCode = 400
		response.Body = "Payment Failed. Please try again later"
		return
	}

	jsonResponse, _ := json.Marshal(&struct{ clientSecret string }{clientSecret})
	response.StatusCode = 200
	response.Body = string(jsonResponse)
	return
}

// ===============================================================================================================================
// END PRIVATE FUNCTIONS
// ===============================================================================================================================

// ===============================================================================================================================
// PUBLIC FUNCTIONS
// ===============================================================================================================================

// Handler function is the entry point for the lambda function
func Handler(request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse) {
	response.StatusCode = 500
	response.Body = "Internal Server Error"
	response.Headers = map[string]string{"Access-Control-Allow-Origin": "*"}

	sess, err := session.NewSession()
	if err != nil {
		fmt.Println(err)
		return
	}
	ddbsvc := dynamodb.New(sess)
	concertsTable := os.Getenv("CONCERTS_TABLE")
	ordersTable := os.Getenv("ORDERS_TABLE")
	stripeSecret := os.Getenv("STRIPE_SECRET")

	if concertsTable == "" || ordersTable == "" || stripeSecret == "" {
		fmt.Println("CONCERTS_TABLE ORDERS_TABLE and STRIPE_SECRET all need to be set as environment variables")
		return
	}
	dynamoHandler := ddbHandler.New(ddbsvc, concertsTable, ordersTable)
	stripeHandler := stripePaymentHandler.New(stripeSecret)

	return processPayment(request, dynamoHandler, stripeHandler)

}

func main() {
	lambda.Start(Handler)
}

// ===============================================================================================================================
// END PUBLIC FUNCTIONS
// ===============================================================================================================================
