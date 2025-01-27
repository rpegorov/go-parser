package db

type Enterprise struct {
	ID             int    `gorm:"primaryKey"`
	EnterpriseID   int    `gorm:"unique;not null"`
	EnterpriseName string `gorm:"not null"`
}

type Site struct {
	ID           int    `gorm:"primaryKey"`
	SiteID       int    `gorm:"unique;not null"`
	SiteName     string `gorm:"not null"`
	EnterpriseID int    `gorm:"not null"`
}

type Department struct {
	ID             int    `gorm:"primaryKey"`
	DepartmentID   int    `gorm:"unique;not null"`
	DepartmentName string `gorm:"not null"`
	SiteID         int    `gorm:"not null"`
}

type Equipment struct {
	ID            int    `gorm:"primaryKey"`
	EquipmentID   int    `gorm:"unique;not null"`
	EquipmentName string `gorm:"not null"`
	DepartmentID  int    `gorm:"not null"`
}

type Indicator struct {
	ID            int    `gorm:"primaryKey"`
	IndicatorID   int    `gorm:"unique;not null"`
	IndicatorName string `gorm:"not null"`
	EquipmentID   int    `gorm:"not null"`
}

type TimeSeries struct {
	ID          int    `gorm:"primaryKey"`
	IndicatorID int    `gorm:"not null"`
	EquipmentID int    `gorm:"not null"`
	DateTime    string `gorm:"not null"`
	Value       string
}

type ExtendedWorkCenter struct {
	ID                    int `gorm:"primaryKey"`
	RecordStartDate       string
	RecordEndDate         string
	ProcessingProgram     string
	EquipmentID           int    `gorm:"not null"`
	MachineStateType      int    `gorm:"not null"`
	DownTimeReasons       int    `gorm:"not null"`
	ReferenceBookReasonID int    `gorm:"not null"`
	ReasonName            string `gorm:"not null"`
	UserName              string `gorm:"not null"`
	OperatorComment       string
	IDRecord              int64  `gorm:"not null"`
	Start                 string `gorm:"not null"`
	End                   string
}
