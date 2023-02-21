package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler"
	"github.com/philomusica/tickets-lambda-get-concerts/lib/databaseHandler/ddbHandler"
	"github.com/philomusica/tickets-lambda-process-payment/lib/emailHandler"
	"github.com/philomusica/tickets-lambda-process-payment/lib/emailHandler/sesEmailHandler"
	"github.com/philomusica/tickets-lambda-process-payment/lib/paymentHandler"
	"github.com/philomusica/tickets-lambda-process-payment/lib/paymentHandler/stripePaymentHandler"
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
		if ol.NumOfFullPrice == nil || ol.NumOfConcessions == nil || ol.ConcertId == "" {
			err = ErrInvalidRequestBody{Message: "order line is missing requirement information"}
			return
		}
	}
	return
}

// processPayment is the main function, taking the AWS events.APIGatewayProxyRequest struct, a DatabaseHandler and PaymentHandler (both interfaces) and response an AWS events.APIGatewayProxyResponse struct
func processPayment(request events.APIGatewayProxyRequest, dbHandler databaseHandler.DatabaseHandler, payHandler paymentHandler.PaymentHandler, emailHandler emailHandler.EmailHandler) (response events.APIGatewayProxyResponse) {
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
		concert, err = dbHandler.GetConcertFromTable(ol.ConcertId)
		if err != nil {
			fmt.Println(err)
			response.StatusCode = 400
			response.Body = err.Error()
			return
		}

		ticketTotal := uint16(*ol.NumOfFullPrice + *ol.NumOfConcessions)
		if concert.AvailableTickets < ticketTotal {
			err = ErrInsufficientAvailableTickets{Message: fmt.Sprintf("Insufficient tickets available for %s\n", concert.Description)}
			fmt.Println(err)
			response.StatusCode = 403
			response.Body = err.Error()
			return
		}

		balance += float32(*ol.NumOfFullPrice)*concert.FullPrice + float32(*ol.NumOfConcessions)*concert.ConcessionPrice
		concerts[ol.ConcertId] = *concert
	}

	// TODO Implement Process function
	err = payHandler.Process(payReq, balance)
	if err != nil {
		fmt.Println(err)
		response.StatusCode = 400
		response.Body = "Payment Failed. Please try again later"
		return
	}

	for _, ol := range payReq.OrderLines {
		// Set default error message
		errMessage := "Payment was successful but we were unable to send a confirmation email with your tickets. Please contact us at https://philomusica.org.uk/contact"
		response.StatusCode = 500
		response.Body = errMessage

		// Update concert table with number of sold tickets
		err := dbHandler.UpdateTicketsSoldInTable(ol.ConcertId, uint16(*ol.NumOfFullPrice+*ol.NumOfConcessions))
		if err != nil {
			fmt.Println(err)
			return
		}

		// Create Order struct
		order := paymentHandler.Order{
			ConcertId:        ol.ConcertId,
			FirstName:        payReq.FirstName,
			LastName:         payReq.LastName,
			Email:            payReq.Email,
			NumOfFullPrice:   *ol.NumOfFullPrice,
			NumOfConcessions: *ol.NumOfConcessions,
		}

		// Add order to orders table
		err = dbHandler.CreateOrderInTable(order)
		if err != nil {
			fmt.Printf("Unable to create order in Orders table: %s\n", err)
			return
		}

		// Generate QR code
		redeemTicketURL := os.Getenv("REDEEM_TICKET_API")
		if redeemTicketURL == "" {
			fmt.Printf("redeemTicketURL not set\n")
			return
		}

		// Generate PDF tickets (injecting QR code)
		attachment := emailHandler.GenerateTicketPDF(order, concerts[order.ConcertId], true, redeemTicketURL)
		if err != nil {
			fmt.Printf("Unable to generate QR code: %s\n", err)
			return
		}

		// Email user with PDF attached
		err = emailHandler.SendEmail(order, attachment)
		if err != nil {
			fmt.Printf("Unable to send email: %s\n", err)
			return
		}
	}
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

	sess, err := session.NewSession()
	if err != nil {
		fmt.Println(err)
		return
	}
	ddbsvc := dynamodb.New(sess)
	sessvc := sesv2.New(sess)
	concertsTable := os.Getenv("CONCERTS_TABLE")
	ordersTable := os.Getenv("ORDERS_TABLE")
	if concertsTable == "" || ordersTable == "" {
		fmt.Println("CONCERT_TABLE and/or ORDERS_TABLE environment variables not set")
		return
	}
	dynamoHandler := ddbHandler.New(ddbsvc, concertsTable, ordersTable)
	stripeHandler := stripePaymentHandler.New()

	senderAddress := os.Getenv("SENDER_ADDRESS")
	if senderAddress == "" {
		fmt.Println("SENDER_ADDRESS env var not set")
		return
	}
	sesHandler := sesEmailHandler.New(sessvc, senderAddress)

	return processPayment(request, dynamoHandler, stripeHandler, sesHandler)

}

func main() {
	lambda.Start(Handler)
}

// ===============================================================================================================================
// END PUBLIC FUNCTIONS
// ===============================================================================================================================
