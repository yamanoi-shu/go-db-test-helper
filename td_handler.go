package main

import (
	"database/sql"

	"gorm.io/gorm"
)

type TDHandler struct {
	db *sql.DB
}

func NewTDHandler(db *sql.DB) (*TDHandler, error) {
	if db.Ping() != nil {
		return nil, DBConnectionErr
	}
	return &TDHandler{db}, nil
}

func NewTDHandlerGorm(db *gorm.DB) (*TDHandler, error) {
	sqlDB, err := db.DB()

	if err != nil {
		return nil, DBConnectionErr
	}

	if sqlDB.Ping() != nil {
		return nil, DBConnectionErr
	}

	return &TDHandler{sqlDB}, nil
}

func (td *TDHandler) CleanUp(tableName string) (int64, error) {
	result, err := td.db.Exec("TRUNCATE table ?", tableName)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
