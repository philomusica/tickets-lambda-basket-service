module github.com/philomusica/tickets-lambda-basket-service

go 1.20

require (
	github.com/aws/aws-lambda-go v1.38.0
	github.com/aws/aws-sdk-go v1.44.213
	github.com/philomusica/tickets-lambda-get-concerts v1.7.0
	github.com/philomusica/tickets-lambda-post-payment v1.0.1
	github.com/stripe/stripe-go/v74 v74.10.0
)

require (
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/philomusica/tickets-lambda-process-payment v1.2.4 // indirect
)
