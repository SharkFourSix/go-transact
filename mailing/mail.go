package mailing

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/mail"
	"time"

	"github.com/SharkFourSix/go-transact/utils"
	"github.com/mhale/smtpd"
	log "github.com/sirupsen/logrus"
)

const (
	APPLICATION_NAME = "go-transact-smtpd"
)

type EmailReceivedHandler func(ip net.Addr, from string, to []string, subject string, data string)
type SourceAddressVerier func(remoteAddr net.Addr, from string, to string) bool

type MailServer struct {
	Address         string
	Handler         EmailReceivedHandler
	SrcAddrVerifier SourceAddressVerier
	UseTLS          bool
	CertificateFile string
	KeyFile         string
	KeyPassphrase   string
	server          *smtpd.Server
}

// Any mail that does not match a defined template will be treated as spam
type SpamMail struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	Body      string
	IpAddress string
	Subject   string
	Email     string
}

type TransactionEmail struct {
	ID         string `gorm:"primaryKey"`
	CreatedAt  time.Time
	Body       string
	IpAddress  string
	Subject    string
	From       string
	Recipients string
}

func (ms *MailServer) Start() error {
	handler := func(origin net.Addr, from string, to []string, data []byte) error {
		msg, err := mail.ReadMessage(bytes.NewReader(data))

		if err != nil {
			log.Errorf("error reading email from address %s. %s", origin.String(), err.Error())
			return err
		}

		body, err := ioutil.ReadAll(msg.Body)
		if err != nil {
			log.Errorf("error reading email from address %s. %s", origin.String(), err.Error())
			return err
		}

		message := string(body[:])
		subject := msg.Header.Get("Subject")

		go ms.Handler(origin, from, to, subject, message)

		return nil
	}

	authHandler := func(remoteAddr net.Addr, mechanism string, username []byte, password []byte, shared []byte) (bool, error) {
		return false, fmt.Errorf("not supported")
	}

	handlerRcpt := func(remoteAddr net.Addr, from string, to string) bool {
		return ms.SrcAddrVerifier(remoteAddr, from, to)
	}

	if ms.Handler == nil {
		panic(fmt.Errorf("a handler is required"))
	}

	if ms.SrcAddrVerifier == nil {
		panic(fmt.Errorf("a source address verifier handler is required"))
	}

	var err error
	ms.server = &smtpd.Server{
		Addr:        ms.Address,
		Appname:     APPLICATION_NAME,
		Handler:     handler,
		HandlerRcpt: handlerRcpt,
		AuthHandler: authHandler,
	}

	if ms.UseTLS {
		if utils.IsStringEmpty(ms.CertificateFile) || utils.IsStringEmpty(ms.KeyFile) {
			return fmt.Errorf("certificate or key file missing")
		}
		if utils.IsStringEmpty(ms.KeyPassphrase) {
			err = ms.server.ConfigureTLS(ms.CertificateFile, ms.KeyFile)
		} else {
			err = ms.server.ConfigureTLSWithPassphrase(ms.CertificateFile, ms.KeyFile, ms.KeyPassphrase)
		}
		if err != nil {
			return fmt.Errorf("error configuring TLS. %v", err)
		}
	}
	if err = ms.server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start SMTP deamon. %v", err)
	}
	return nil
}

func (ms *MailServer) Shutdown(ctx context.Context) error {
	return ms.server.Shutdown(ctx)
}
