package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/twinj/uuid"

	"github.com/SharkFourSix/go-transact/config"
	"github.com/SharkFourSix/go-transact/mailing"
	"github.com/SharkFourSix/go-transact/messaging"
	"github.com/SharkFourSix/go-transact/persistence"
	"github.com/SharkFourSix/go-transact/transaction"
	"github.com/SharkFourSix/go-transact/utils"
	"github.com/devfacet/gocmd"
)

const (
	NAME         = "go-transact"
	VERSION      = "1.0"
	DESCRIPTION  = "Bank transaction notification to action service"
	BUSY_TIMEOUT = 5000
)

func main() {

	flags := struct {
		Version    bool   `short:"v" long:"version" description:"Display version"`
		Help       bool   `short:"h" long:"help" description:"Show help"`
		Verbose    bool   `short:"x" long:"verbose" description:"Set verbose to on"`
		ConfigFile string `short:"c" long:"config-file" description:"Path to configuration file"`
	}{}

	var (
		verbose    bool
		configFile string
		mailServer *mailing.MailServer
		exitStatus int = 1
	)

	_, _ = gocmd.HandleFlag("Help", func(cmd *gocmd.Cmd, args []string) error {
		cmd.PrintUsage()
		return nil
	})

	_, _ = gocmd.HandleFlag("ConfigFile", func(cmd *gocmd.Cmd, args []string) error {
		configFile = args[0]
		if !utils.FileExists(configFile) {
			return fmt.Errorf("config file does not exist: '%s'", configFile)
		}
		return nil
	})

	_, _ = gocmd.HandleFlag("Verbose", func(cmd *gocmd.Cmd, args []string) error {
		verbose = true
		return nil
	})

	_, _ = gocmd.New(gocmd.Options{
		Name:        NAME,
		Description: DESCRIPTION,
		Version:     fmt.Sprintf("%s v%s", NAME, VERSION),
		Flags:       &flags,
		ConfigType:  gocmd.ConfigTypeAuto,
		AutoHelp:    true,
	})

	defer func() {
		log.Infof("exit %d", exitStatus)
		os.Exit(exitStatus)
	}()

	if utils.IsStringEmpty(configFile) {
		fmt.Println(errors.New("missing parameter '--config-file'. Use '--help' for more information"))
		return
	}

	err := config.LoadConfigs(configFile, verbose)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Debug("Opening database")
	if err := persistence.Initialize(BUSY_TIMEOUT); err != nil {
		log.Errorf("Error initializing database. %s\n", err.Error())
		return
	}
	defer persistence.Cleanup()

	log.Debug("Applying migrations")
	if err := persistence.Migrate(&transaction.Transaction{}, &messaging.TransactionNotification{},
		&mailing.SpamMail{}, &mailing.TransactionEmail{}); err != nil {
		log.Errorf("Error running database migrations. %s\n", err.Error())
		return
	}

	mailboxVerifier := func(remoteAddr net.Addr, from string, to string) bool {
		parts := strings.Split(to, "@")
		if len(parts) <= 1 {
			log.Debugf("rejected host %s because of malformed mailbox name.", remoteAddr.String())
			return false
		}
		exists := config.MailBoxExists(parts[0])
		log.Debugf("%s: mailbox exists %s: %t", remoteAddr.String(), parts[0], exists)
		return exists
	}

	handler := func(ip net.Addr, from string, to []string, subject string, data string) {
		log.Debugf("Got email from ip %s, sender %s", ip.String(), from)

		template := config.GetTemplateByEmail(from)

		if template == nil {
			log.Warnf("sender %s did not match any template. Email will be stored in spam", from)
			spam := mailing.SpamMail{
				ID:        uuid.NewV4().String(),
				Body:      data,
				Email:     from,
				Subject:   subject,
				IpAddress: ip.String(),
				CreatedAt: time.Now(),
			}
			if err := persistence.Save(&spam); err != nil {
				log.Errorf("failed to save spam mail from %s, %s", ip.String(), from)
			}
			return
		}

		email := mailing.TransactionEmail{
			ID:         uuid.NewV4().String(),
			CreatedAt:  time.Now(),
			Body:       data,
			IpAddress:  ip.String(),
			Subject:    subject,
			From:       from,
			Recipients: strings.Join(to, ","),
		}

		log.Debugf("saving transaction email [server=%s, sender=%s]", ip.String(), from)
		if err := persistence.Save(&email); err != nil {
			log.Errorf("failed to save mail from [server=%s, sender=%s] for template %s. %s",
				from, ip.String(), template.TemplateName, err.Error())
		}

		log.Debugf("parsing transaaction from %s using template %s.", from, template.TemplateName)
		transaction, err := transaction.ParseTransaction(data, template)
		if err != nil {
			log.Errorf("failed to parse transaction. %s", err.Error())
			return
		}

		if err := persistence.Save(transaction); err != nil {
			log.Errorf("failed to save transaction. %s", err.Error())
		}

		callback := messaging.NotificationData{
			CreatedAt:              time.Now(),
			TemplateName:           transaction.TemplateName,
			Date:                   transaction.Date,
			Amount:                 transaction.Amount,
			Currency:               transaction.Currency,
			AccountNumber:          transaction.AccountNumber,
			VendorReferenceId:      transaction.VendorReferenceId,
			TransactionReferenceId: transaction.TransactionReferenceId,
		}
		notificationLog := messaging.TransactionNotification{
			FromEmail:    from,
			Sent:         false,
			CreatedAt:    time.Now(),
			ID:           uuid.NewV4().String(),
			TemplateName: transaction.TemplateName,
			Url:          config.GetConfiguration().Callback.ForwardURL,
		}

		if err := notificationLog.Post(config.GetConfiguration().Callback.ForwardToken, &callback); err != nil {
			log.Errorf("failure posting notification for transaction from %s. %s", from, err.Error())
		}

		if err := persistence.Save(&callback); err != nil {
			log.Errorf("failure saving notification. %s. response was %s", err.Error(), notificationLog.StatusText)
		}
	}

	mailServer = &mailing.MailServer{
		Address:         config.GetConfiguration().Server.Address,
		UseTLS:          config.GetConfiguration().Server.UserTls,
		CertificateFile: config.GetConfiguration().Server.CertificateFile,
		KeyFile:         config.GetConfiguration().Server.KeyFile,
		KeyPassphrase:   config.GetConfiguration().Server.KeyPassphrase,
		Handler:         handler,
		SrcAddrVerifier: mailboxVerifier,
	}

	exitChannel := make(chan int)

	go func() {
		log.Debug("starting mail server...")
		defer func() {
			log.Info("shutting down daemon...")
			if err := mailServer.Shutdown(context.Background()); err != nil {
				log.Errorf("error during server shutdown %s", err.Error())
			}
		}()
		if err := mailServer.Start(); err != nil {
			exitChannel <- 1
			log.Error(err)
		}
	}()

	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

		fmt.Println("Press Ctrl+C to stop")

		signal := <-signalChannel

		log.Info(signal.String())
		fmt.Printf("\n%s\n", signal.String())

		exitChannel <- 0
	}()

	log.Debug("Up and running. Waiting for signals")

	exitStatus = <-exitChannel
}
