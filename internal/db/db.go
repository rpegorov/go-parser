package db

import (
	"log"

	// "github.com/rpegorov/go-parser/internal/utils"
	"github.com/rpegorov/go-parser/internal/utils"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Databases struct {
	PostgresDB   *gorm.DB
	ClickHouseDB *gorm.DB
}

func Init() *Databases {
	postgresDB := initPostgres()
	clickHouseDB := initClickHouse()

	return &Databases{
		PostgresDB:   postgresDB,
		ClickHouseDB: clickHouseDB,
	}
}

func initPostgres() *gorm.DB {
	dbURL := utils.GoDotEnvVariable("DB_URL")

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}

	err = db.AutoMigrate(
		&Enterprise{},
		&Site{},
		&Department{},
		&Equipment{},
		&Indicator{},
		&ExtendedWorkCenter{},
	)
	if err != nil {
		log.Fatalf("Ошибка миграции PostgreSQL: %v", err)
	}

	log.Println("Успешное подключение к PostgreSQL")
	return db
}

func initClickHouse() *gorm.DB {
	dsn := utils.GoDotEnvVariable("CLICK_URL")
	db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	log.Println("Успешное подключение к ClickHouse")
	db.AutoMigrate(&TimeSeries{})
	return db
}
