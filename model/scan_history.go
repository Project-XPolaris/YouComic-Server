package model

import "gorm.io/gorm"

type ScanHistory struct {
	gorm.Model
	LibraryId  uint `gorm:"index"`
	Total      int64
	ErrorCount int
	Status     string
	StartedAt  int64
	FinishedAt int64
	Errors     string `gorm:"type:text"`
}
