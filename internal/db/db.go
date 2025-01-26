package db

import (
	"log"

	"github.com/rpegorov/go-parser/internal/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init() *gorm.DB {
	dsn := utils.GoDotEnvVariable("DB_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	db.AutoMigrate(
		&Enterprise{},
		&Site{},
		&Department{},
		&Equipment{},
		&Indicator{},
		&TimeSeries{},
		&ExtendedWorkCenter{})

	return db
}
