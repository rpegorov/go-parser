package migrations

import (
	"log"

	"gorm.io/gorm"
)

func MigrateClickHouse(db *gorm.DB) {
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS time_series
    (id String, indicator_id UInt32, equipment_id UInt32, date_time DateTime64(3, 'Europe/Moscow'), value String)
    ENGINE MergeTree()
    ORDER BY (date_time)
  `

	if err := db.Exec(createTableQuery); err != nil {
		log.Fatalf("Error creating ClickHouse table: %v", err)
	}

	log.Println("ClickHouse migration completed successfully!")
}
