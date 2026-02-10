package database

import (
	"database/sql"

	"github.com/kyawphyothu/sana/config"
	_ "github.com/mattn/go-sqlite3"
)

func NewDB(config *config.Config) (*sql.DB, error) {
	db, err := sql.Open(config.DBType, config.DBPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}
