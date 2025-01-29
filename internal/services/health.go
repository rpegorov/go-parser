package services

import "gorm.io/gorm"

type HealthService interface {
	CheckHealth() (map[string]string, error)
}

type HealthServiceImpl struct {
	PostgresDB   *gorm.DB
	ClickHouseDB *gorm.DB
}

func NewHealthService(postgres *gorm.DB, clickhouse *gorm.DB) *HealthServiceImpl {
	return &HealthServiceImpl{
		PostgresDB:   postgres,
		ClickHouseDB: clickhouse,
	}
}

func (s *HealthServiceImpl) CheckHealth() (map[string]string, error) {
	var postgresStatus string
	if s.PostgresDB != nil {
		pgDB, err := s.PostgresDB.DB()
		if err != nil || pgDB.Ping() != nil {
			postgresStatus = "PostgresDB: Unreachable"
		} else {
			postgresStatus = "PostgresDB: Connected"
		}
	} else {
		postgresStatus = "PostgresDB: Not Configured"
	}

	var clickHouseStatus string
	if s.ClickHouseDB != nil {
		chDB, err := s.ClickHouseDB.DB()
		if err != nil || chDB.Ping() != nil {
			clickHouseStatus = "ClickHouseDB: Unreachable"
		} else {
			clickHouseStatus = "ClickHouseDB: Connected"
		}
	} else {
		clickHouseStatus = "ClickHouseDB:	: Not Configured"
	}

	status := map[string]string{
		"PostgresDB":   postgresStatus,
		"ClickHouseDB": clickHouseStatus,
	}

	return status, nil
}
