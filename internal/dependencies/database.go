package dependencies

import (
	"fmt"
	"log"

	"github.com/cynxees/ra-server/internal/dependencies/config"
	"github.com/cynxees/ra-server/internal/model/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type DatabaseClient struct {
	DB *gorm.DB
}

func NewDatabaseClient() (*DatabaseClient, error) {
	// Construct the DSN (Data Source Name)
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
		config.Config.Database.Username,
		config.Config.Database.Password,
		config.Config.Database.Host,
		config.Config.Database.Port,
		config.Config.Database.Database,
	)

	// Open a connection with GORM using the MySQL driver
	db, err := gorm.Open(mysql.Open(dataSourceName), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Enable GORM's built-in logging
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		QueryFields: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Check the connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get generic database object: %w", err)
	}
	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseClient{DB: db}, nil
}

func (client *DatabaseClient) Close() error {
	sqlDB, err := client.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get generic database object: %w", err)
	}
	return sqlDB.Close()
}

func (client *DatabaseClient) RunMigrations() error {
	log.Println("Running database migrations")
	err := client.DB.AutoMigrate(&entity.VirtualMachine{})
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}
