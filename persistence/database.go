package persistence

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var databaseHandle *gorm.DB

func Initialize(timeout int) error {
	var dsn = fmt.Sprintf("transactions.db?_busy_timeout=%d&_journal_mode=WAL", timeout)
	var err error

	databaseHandle, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	return nil
}

func Cleanup() {
	log.Info("Closing database")
	db, err := databaseHandle.DB()
	if err != nil {
		log.Warnf("Error getting database handle. %s", err.Error())
		return
	}
	if err := db.Close(); err != nil {
		log.Warnf("Error getting database handle. %s", err.Error())
		return
	}
	log.Info("Database closed")
}

func Migrate(models ...interface{}) error {
	return databaseHandle.AutoMigrate(models...)
}

func saveModel(model interface{}) *gorm.DB {
	return databaseHandle.Create(model)
}

func Save(model interface{}) error {
	return saveModel(model).Error
}
