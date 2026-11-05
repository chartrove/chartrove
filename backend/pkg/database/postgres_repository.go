package database

import (
	"fmt"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/config"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/event_bus"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newPostgresRepository(appConfig config.Interface, globalLogger logrus.FieldLogger, eventBus event_bus.Interface) (DatabaseRepository, error) {
	//backgroundContext := context.Background()

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Gorm/PostgreSQL setup
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	globalLogger.Info("Trying to connect to postgres database")
	dsn := appConfig.GetString("database.location")

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Do not emit SQL or driver-detail traces containing clinical data or
		// credentials. Application lifecycle diagnostics remain independent.
		Logger:                                   logger.Discard,
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database! - %v", err)
	}
	globalLogger.Info("Successfully connected to postgres database")

	fastenRepo := GormRepository{
		AppConfig:  appConfig,
		Logger:     globalLogger,
		GormClient: database,
		EventBus:   eventBus,
	}

	//TODO: automigrate for now, this should be replaced with a migration tool once the DB has stabilized.
	err = fastenRepo.Migrate()
	if err != nil {
		_ = fastenRepo.Close()
		return nil, err
	}

	//fail any Locked jobs. This is necessary because the job may have been locked by a process that was killed.
	err = fastenRepo.CancelAllLockedBackgroundJobsAndFail()
	if err != nil {
		_ = fastenRepo.Close()
		return nil, err
	}

	return &fastenRepo, nil
}
