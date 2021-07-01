package database

import (
	"pointers/pointers/config"

	"github.com/syndtr/goleveldb/leveldb"
)

type Database struct {
	db *leveldb.DB
}

func NewDatabase(config *config.DatabaseConfig) (*Database, error) {
	db, err := leveldb.OpenFile(config.Path, nil)
	if err != nil {
		return nil, err
	}
	return &Database{
		db: db,
	}, nil
}

func (db *Database) Close() {
	db.db.Close()
}
