package messaging

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/SharkFourSix/go-transact/persistence"
	log "github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
)

func TestTransactionNotificationPersistence(t *testing.T) {
	var notification = TransactionNotification{
		CreatedAt:    time.Now(),
		Url:          "http://localhost:8080/transaction_callback",
		Data:         `{"transaction":{"amount":"56,000.00","foo":"bar"}}`,
		Sent:         true,
		StatusText:   "Successfully posted",
		ResponseText: "200 OK",
		FromEmail:    "bankemail@host.tld",
		TemplateName: "Template Name",
		ID:           uuid.NewV4().String(),
	}

	if err := persistence.Initialize(5000); err != nil {
		t.Fatal(err)
	}

	if err := persistence.Migrate(&TransactionNotification{}); err != nil {
		t.Fatal(err)
	}

	if err := persistence.Save(&notification); err != nil {
		t.Fatal(err)
	}

	defer func() {
		persistence.Cleanup()
	}()
}

func TestTransactionNotificationPost(t *testing.T) {
	var notification = TransactionNotification{
		CreatedAt:    time.Now(),
		Url:          "http://localhost:45900/callback_test",
		Sent:         true,
		StatusText:   "Successfully posted",
		ResponseText: "200 OK",
		FromEmail:    "bankemail@host.tld",
		TemplateName: "Template Name",
		ID:           uuid.NewV4().String(),
	}

	var data = NotificationData{
		CreatedAt:              time.Now(),
		TemplateName:           "Test template",
		Date:                   "20210101",
		Amount:                 "50,342.00",
		Currency:               "MWK",
		AccountNumber:          "1239874",
		VendorReferenceId:      "VRIF0XA65FE2",
		TransactionReferenceId: "FT9832747842",
	}

	log.SetLevel(log.DebugLevel)

	poster := func(server *http.Server, c chan error) {
		if err := notification.Post("token12345", &data); err != nil {
			c <- err
			server.Shutdown(context.Background())
			return
		}
		c <- nil
		server.Shutdown(context.Background())
	}

	var server = &http.Server{Addr: ":45900", Handler: nil}
	errc := make(chan error)

	http.HandleFunc("/callback_test", func(w http.ResponseWriter, r *http.Request) {
		var buff []byte = make([]byte, 1024)
		log.Debugf("Received %s request from %s", r.Method, r.Host)

		count, _ := r.Body.Read(buff)

		log.Debugln(string(buff[:count]))

		server.Close()
		//server.Shutdown(context.Background())
		errc <- nil
	})

	go poster(server, errc)

	server.ListenAndServe()

	if err := <-errc; err != nil {
		t.Fatal(err)
	}

}
