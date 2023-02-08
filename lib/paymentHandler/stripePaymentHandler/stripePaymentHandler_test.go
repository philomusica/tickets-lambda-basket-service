package stripePaymentHandler

import (
	"fmt"
	"github.com/philomusica/tickets-lambda-process-payment/lib/paymentHandler"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	rc := m.Run()
	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		fmt.Println(c)
		if c < 0.9 {
			fmt.Printf("Tests passed but coverage was below %d%%\n", int(c*100))
			rc = -1
		}
	}
	os.Exit(rc)
}

// ===============================================================================================================================
// PROCESS TESTS
// ===============================================================================================================================

func TestStripePaymentHandlerFails(t *testing.T) {
	stripeHandler := New()

	payReq := paymentHandler.PaymentRequest{}
	var balance float32 = 40.0
	err := stripeHandler.Process(payReq, balance)
	expectedErr, ok := err.(paymentHandler.ErrPaymentFailed)

	if !ok {
		t.Errorf("Expected err %T, got %T\n", expectedErr, err)
	}
}

// ===============================================================================================================================
// END PROCESS TESTS
// ===============================================================================================================================
