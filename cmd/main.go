package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/philomusica/tickets-lambda-utils/lib/databaseHandler"
	"github.com/philomusica/tickets-lambda-utils/lib/databaseHandler/ddbHandler"
	"github.com/philomusica/tickets-lambda-utils/lib/paymentHandler"
	"github.com/philomusica/tickets-lambda-utils/lib/paymentHandler/stripePaymentHandler"
)

const DEFAULT_JSON_RESPONSE string = `{"error": "payment failed"}`

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
	var tmpValue float64
	transactionFeePercentageStr := os.Getenv("TRANSACTION_FEE_PERCENTAGE")
	transactionFeeFlatRateStr := os.Getenv("TRANSACTION_FEE_FLAT_RATE")

	var transactionFeePercentage, transactionFeeFlatRate float32 = 0.0, 0.0

	if transactionFeePercentageStr != "" {
		tmpValue, err = strconv.ParseFloat(transactionFeePercentageStr, 32)
		transactionFeePercentage = float32(tmpValue)
	}
	if transactionFeeFlatRateStr != "" {
		tmpValue, err = strconv.ParseFloat(transactionFeeFlatRateStr, 32)
		transactionFeePercentage = float32(tmpValue)
	}

	if err != nil {
		fmt.Println("Issue parsing transaction fee amounts", err)
		return
	}

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

		err = dbHandler.ReformatDateTimeAndTickets(concert)
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
	balance = balance*(transactionFeePercentage/100+1) + transactionFeeFlatRate

	orderReference := dbHandler.GenerateOrderReference(4)
	for _, ol := range payReq.OrderLines {
		// Set default error message
		errMessage := "Internal Server Error"
		response.StatusCode = 500
		response.Body = errMessage

		// Create Order struct
		order := paymentHandler.Order{
			OrderReference:   orderReference,
			ConcertID:        ol.ConcertID,
			FirstName:        payReq.FirstName,
			LastName:         payReq.LastName,
			Email:            payReq.Email,
			NumOfFullPrice:   *ol.NumOfFullPrice,
			NumOfConcessions: *ol.NumOfConcessions,
			AdditionalFields: ol.AdditionalFields,
			OrderStatus:           "pending",
		}

		// Add order to orders table
		err = dbHandler.CreateOrderInTable(order)
		if err != nil {
			fmt.Printf("Unable to create order in Orders table: %s\n", err)
			return
		}
	}

	clientSecret, err := payHandler.Process(balance, orderReference)
	if err != nil {
		fmt.Println(err)
		response.StatusCode = 400
		response.Body = "Payment Failed. Please try again later"
		return
	}

	jsonResponse, _ := json.Marshal(struct {
		ClientSecret string `json:"clientSecret"`
	}{clientSecret})
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
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{
		Body:       DEFAULT_JSON_RESPONSE,
		StatusCode: 500,
		Headers:    map[string]string{"Access-Control-Allow-Origin": "https://philomusica.org.uk"},
	}

	sess, err := session.NewSession()
	if err != nil {
		fmt.Println(err)
		return response, nil
	}
	ddbsvc := dynamodb.New(sess)
	concertsTable := os.Getenv("CONCERTS_TABLE")
	ordersTable := os.Getenv("ORDERS_TABLE")
	stripeSecret := os.Getenv("STRIPE_SECRET")

	if concertsTable == "" || ordersTable == "" || stripeSecret == "" {
		fmt.Println("CONCERTS_TABLE ORDERS_TABLE and STRIPE_SECRET all need to be set as environment variables")
		return response, nil
	}
	dynamoHandler := ddbHandler.New(ddbsvc, concertsTable, ordersTable)
	stripeHandler := stripePaymentHandler.New(stripeSecret)

	return processPayment(request, dynamoHandler, stripeHandler), nil

}

func main() {
	lambda.Start(Handler)
}

// ===============================================================================================================================
// END PUBLIC FUNCTIONS
// ===============================================================================================================================
