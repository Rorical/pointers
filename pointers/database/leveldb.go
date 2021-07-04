package database

import (
	"pointers/pointers/config"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type Database struct {
	db   *leveldb.DB
	lock *sync.RWMutex
}

func NewDatabase(config *config.DatabaseConfig) (*Database, error) {
	opts := &opt.Options{
		Filter: filter.NewBloomFilter(10),
	}
	db, err := leveldb.OpenFile(config.Path, opts)
	if err != nil {
		return nil, err
	}
	return &Database{
		db:   db,
		lock: new(sync.RWMutex),
	}, nil
}

func (db *Database) Get(key string) (*[]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	val, err := db.db.Get(S2B(key), nil)
	return &val, err
}

func (db *Database) Set(key string, value *[]byte) error {
	db.lock.RLock()
	defer db.lock.RUnlock()
	err := db.db.Put(S2B(key), *value, &opt.WriteOptions{Sync: true})
	return err
}

func (db *Database) Del(key string) error {
	db.lock.RLock()
	defer db.lock.RUnlock()
	err := db.db.Delete(S2B(key), &opt.WriteOptions{Sync: true})
	return err
}

func (db *Database) Has(key string) (bool, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	val, err := db.db.Has(S2B(key), nil)
	return val, err
}

func (db *Database) Close() {
	db.db.Close()
}
