package transaction

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/SharkFourSix/go-transact/persistence"
	"github.com/SharkFourSix/go-transact/utils"
	log "github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
)

const (
	NBM_MESSAGE = `Dear MR MR JOHN DOE,
	We advise that your account number 12345678 has been credited with MWK20,000.00 on 20220505.
	Description: 98324HAZ123P003.
	Reference: FT12345K1234\BNK.
	Current Balance: 1,098,724.75.
	Available Balance: 1,093,724.75.
	Cleared Balance: 1,098,724.75.
	Thank you for banking with us.`
)

func TestTransaction(t *testing.T) {
	var template = &TransactionTemplate{
		TemplateName:                  "National Bank Of Malawi",
		Email:                         "n/a",
		DatePattern:                   "on (?P<date>[0-9]{8})",
		AmountPattern:                 "(?P<amount>[0-9,.]{3,18}) on ",
		CurrencyPattern:               "with (?P<currency>[A-Z]{3})",
		AccountNumberPattern:          "account number (?P<accountNumber>[0-9]+)",
		VendorReferenceIdPattern:      `Description: (?P<vendorReferenceId>[0-9A-Za-z]{1,255})\.$`,
		TransactionReferenceIdPattern: `Reference: (?P<transactionReferenceId>FT[0-9A-Z]+\\BNK)\.$`,
	}

	log.SetLevel(log.DebugLevel)

	var tx, err = ParseTransaction(NBM_MESSAGE, template)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if utils.IsStringEmpty(tx.Amount) {
		t.Fatal("Amount did not match")
	}

	if utils.IsStringEmpty(tx.Date) {
		t.Fatal("Date did not match")
	}

	if utils.IsStringEmpty(tx.AccountNumber) {
		t.Fatalf("Account number did not match")
	}

	if utils.IsStringEmpty(tx.Currency) {
		t.Fatalf("Currency did not match")
	}

	if utils.IsStringEmpty(tx.TransactionReferenceId) {
		t.Fatalf("Transaction reference id did not match")
	}

	if utils.IsStringEmpty(tx.VendorReferenceId) {
		t.Fatalf("Vendor reference id did not match")
	}

	// Only works if -test.v is specified
	out, _ := json.Marshal(tx)
	t.Logf("Transaction: %s\n", out)
}

func TestSaveTransaction(t *testing.T) {
	var transaction = Transaction{
		ID:                     uuid.NewV4().String(),
		CreatedAt:              time.Now(),
		Date:                   "20210101",
		TemplateName:           "Sample Template Name",
		Currency:               "USD",
		Amount:                 "50,000.00",
		AccountNumber:          "500674534453",
		VendorReferenceId:      "VRF1234567890",
		TransactionReferenceId: "FT123456789123",
	}

	if err := persistence.Initialize(5000); err != nil {
		t.Fatal(err)
	}

	if err := persistence.Migrate(&Transaction{}); err != nil {
		t.Fatal(err)
	}

	if err := persistence.Save(&transaction); err != nil {
		t.Fatal(err)
	}

	defer func() {
		persistence.Cleanup()
	}()
}
