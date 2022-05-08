package transaction

import (
	"fmt"
	"log"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/twinj/uuid"

	"github.com/SharkFourSix/go-transact/utils"
)

/*
	Represents a credit transaction
*/
type Transaction struct {
	ID                     string `gorm:"primaryKey"`
	CreatedAt              time.Time
	TemplateName           string
	Date                   string
	Amount                 string
	Currency               string
	AccountNumber          string
	VendorReferenceId      string
	TransactionReferenceId string
}

/* Template used for parsing transactions from messages */
type TransactionTemplate struct {
	Email                         string `yaml:"email"`
	TemplateName                  string `yaml:"name"`
	DatePattern                   string `yaml:"datePattern"`
	AmountPattern                 string `yaml:"amountPattern"`
	CurrencyPattern               string `yaml:"currencyPattern"`
	AccountNumberPattern          string `yaml:"accountNumberPattern"`
	VendorReferenceIdPattern      string `yaml:"vendorReferenceIdPattern"`
	TransactionReferenceIdPattern string `yaml:"transactionReferenceIdPattern"`
}

func ParseTransaction(text string, template *TransactionTemplate) (*Transaction, error) {
	var (
		getTransactionField = func(value *string, name string, pattern string, required bool) error {
			var match *regexp.Match

			defer func() {
				if err := recover(); err != nil {
					log.Fatalf("error when parsing transaction: %s", err)
				}
			}()

			matcher, err := regexp.Compile(pattern, regexp.Multiline|regexp.RE2)
			if err != nil {
				return parserError(template, err.Error())
			}
			matcher.MatchTimeout = 5000
			if match, err = matcher.FindStringMatch(text); err != nil || match == nil {
				if err != nil {
					return parserError(template, err.Error())
				}
				return parserError(template, fmt.Sprintf("pattern for group %s did not match.", name))
			}
			if group := match.GroupByName(name); group != nil {
				if utils.IsStringEmpty(group.String()) {
					if required {
						return parserError(template, fmt.Sprintf("missing required field %s", name))
					}
					*value = ""
					return nil
				}
				*value = group.String()
				return nil
			}
			if required {
				return parserError(template, fmt.Sprintf("missing required field %s", name))
			}
			*value = ""
			return nil
		}
	)

	var transaction = &Transaction{
		ID:           uuid.NewV4().String(),
		TemplateName: template.TemplateName,
		CreatedAt:    time.Now(),
	}

	if err := getTransactionField(&transaction.VendorReferenceId, "vendorReferenceId", template.VendorReferenceIdPattern, true); err != nil {
		return nil, parserError(template, err.Error())
	}

	if err := getTransactionField(&transaction.Amount, "amount", template.AmountPattern, true); err != nil {
		return nil, parserError(template, "Missing amount value")
	}

	if err := getTransactionField(&transaction.Date, "date", template.DatePattern, true); err != nil {
		return nil, parserError(template, err.Error())
	}

	if err := getTransactionField(&transaction.TransactionReferenceId, "transactionReferenceId", template.TransactionReferenceIdPattern, false); err != nil {
		return nil, parserError(template, err.Error())
	}

	if err := getTransactionField(&transaction.AccountNumber, "accountNumber", template.AccountNumberPattern, false); err != nil {
		return nil, parserError(template, err.Error())
	}

	if err := getTransactionField(&transaction.Currency, "currency", template.CurrencyPattern, false); err != nil {
		return nil, parserError(template, err.Error())
	}

	return transaction, nil
}

func parserError(template *TransactionTemplate, message string) error {
	return fmt.Errorf("error parsing transaction[%s]: %s", template.TemplateName, message)
}
