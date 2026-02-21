package database

import (
	"database/sql"

	"github.com/kyawphyothu/sana/config"
	_ "modernc.org/sqlite"
)

func NewDB(config *config.Config) (*sql.DB, error) {
	db, err := sql.Open(config.DBType, config.DBPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}
