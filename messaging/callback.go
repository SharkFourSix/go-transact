package messaging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/twinj/uuid"
)

const (
	USER_AGENT_VERSION           = 1
	USER_AGENT_STRING            = "go-transact"
	HTTP_REQUEST_TIMEOUT_SECONDS = 15
)

type TransactionNotification struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	Url       string
	Data      string
	// True = Successfully sent, False = Failure sending
	Sent         bool
	StatusText   string
	ResponseText string
	FromEmail    string
	TemplateName string
}

type NotificationData struct {
	CreatedAt              time.Time
	TemplateName           string
	Date                   string
	Amount                 string
	Currency               string
	AccountNumber          string
	VendorReferenceId      string
	TransactionReferenceId string
}

func NewTransactionNotification() *TransactionNotification {
	return &TransactionNotification{
		ID:        uuid.NewV4().String(),
		CreatedAt: time.Now(),
	}
}

// Post Attempt to send this notification to the specified callback url.
// 	The notification will automatically be updated with status results and response data.
func (n *TransactionNotification) Post(token string, data *NotificationData) error {
	var err error
	var body []byte

	setStatus := func(sent bool, status string, response string) {
		n.Sent = sent
		n.StatusText = status
		n.ResponseText = response
	}

	if body, err = json.Marshal(&data); err != nil {
		err = fmt.Errorf("failure serializing request data %s", err)
		setStatus(false, err.Error(), "")
		return err
	}

	n.Data = string(body[:])

	request, err := http.NewRequest("POST", n.Url, bytes.NewBuffer(body))
	if err != nil {
		err = fmt.Errorf("failure creating request %s", err)
		setStatus(false, err.Error(), "")
		return err
	}

	request.Header.Set("X-Go-Transact-Token", token)
	request.Header.Set("Date", time.Now().UTC().String())
	request.Header.Set("User-Agent", fmt.Sprintf("%s/%d", USER_AGENT_STRING, USER_AGENT_VERSION))

	client := &http.Client{
		Timeout:       time.Second * HTTP_REQUEST_TIMEOUT_SECONDS,
		CheckRedirect: http.DefaultClient.CheckRedirect,
	}

	response, err := client.Do(request)
	if err != nil {
		err = fmt.Errorf("failure sending request to %s. %s", n.Url, err)
		setStatus(false, err.Error(), "")
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		setStatus(true, "Callback posted", "200 OK")
		return nil
	} else {
		err = fmt.Errorf("server returned %d", response.StatusCode)
		// use strings.Join instead of Sptrintf for safety
		setStatus(false, err.Error(), strings.Join([]string{strconv.Itoa(response.StatusCode), response.Status}, " "))
		return err
	}
}
