module github.com/philomusica/tickets-lambda-basket-service

go 1.20

require (
	github.com/aws/aws-lambda-go v1.40.0
	github.com/aws/aws-sdk-go v1.44.246
	github.com/philomusica/tickets-lambda-get-concerts v1.7.12
	github.com/philomusica/tickets-lambda-post-payment v1.0.1
	github.com/stripe/stripe-go/v74 v74.15.0
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
