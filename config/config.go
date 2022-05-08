package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	"github.com/SharkFourSix/go-transact/transaction"
	"github.com/SharkFourSix/go-transact/utils"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Templates []transaction.TransactionTemplate `yaml:"templates"`
	Callback  struct {
		ForwardURL   string `yaml:"url"`
		ForwardToken string `yaml:"token"`
	}
	Server struct {
		Address         string   `yaml:"address"`
		UserTls         bool     `yaml:"useTls"`
		CertificateFile string   `yaml:"certFile"`
		KeyFile         string   `yaml:"keyFile"`
		KeyPassphrase   string   `yaml:"keyPassphrase"`
		Mailboxes       []string `yaml:"mailboxes"`
	}
	Log struct {
		LogLevel   string `yaml:"level"`
		LogFile    string `yaml:"file"`
		FormatJson bool   `yaml:"json"`
	}
}

var configuration Config

func (c *Config) parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}

func (cfg *Config) prepareLogger() error {
	var (
		writer            io.Writer = os.Stderr
		logLevel          log.Level = log.WarnLevel
		callerPrettyfiler           = func(f *runtime.Frame) (function string, file string) {
			directory, _ := os.Getwd()
			return "", fmt.Sprintf("%s:%d", strings.TrimPrefix(f.File, directory), f.Line)
		}
	)

	err := logLevel.UnmarshalText([]byte(cfg.Log.LogLevel))
	if err != nil {
		return fmt.Errorf("invalid log level %s. %s", cfg.Log.LogLevel, err.Error())
	}

	if !utils.IsStringEmpty(cfg.Log.LogFile) {
		writer = &lumberjack.Logger{
			Filename: cfg.Log.LogFile,
			MaxSize:  100,
			Compress: true,
		}
	}

	if cfg.Log.FormatJson {
		log.SetFormatter(&log.JSONFormatter{
			CallerPrettyfier: callerPrettyfiler,
			FieldMap: log.FieldMap{
				logrus.FieldKeyFile: "caller",
			},
		})
	} else {
		log.SetFormatter(&log.TextFormatter{
			CallerPrettyfier: callerPrettyfiler,
			FieldMap: log.FieldMap{
				logrus.FieldKeyFile: "caller",
			},
		})
	}

	log.SetReportCaller(true)
	log.SetOutput(writer)
	log.SetLevel(logLevel)
	return nil
}

func LoadConfigs(file string, verbosity bool) error {

	if verbosity {
		fmt.Printf("Loading config file %s\n...", file)
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("error reading configuration file %s. %s", file, err.Error())
	}

	if err := configuration.parse(data); err != nil {
		return fmt.Errorf("error parsing configuration file %s. %s", file, err.Error())
	}

	if err := configuration.prepareLogger(); err != nil {
		return err
	}

	return nil
}

func GetTemplates() []transaction.TransactionTemplate {
	return configuration.Templates
}

func GetConfiguration() Config {
	return configuration
}

func GetTemplateByEmail(email string) *transaction.TransactionTemplate {
	for _, tpl := range configuration.Templates {
		if strings.EqualFold(tpl.Email, email) {
			return &tpl
		}
	}
	return nil
}

func MailBoxExists(mailbox string) bool {
	for _, mb := range configuration.Server.Mailboxes {
		if strings.EqualFold(mb, mailbox) {
			return true
		}
	}
	return false
}
