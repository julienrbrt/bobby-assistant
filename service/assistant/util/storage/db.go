package storage

import (
	"fmt"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/pebble-dev/bobby-assistant/service/assistant/config"
)

var onceDB sync.Once
var sharedDB *gorm.DB

func GetDB() *gorm.DB {
	onceDB.Do(func() {
		db, err := gorm.Open(sqlite.Open(config.GetConfig().DBPath), &gorm.Config{})
		if err != nil {
			panic(fmt.Errorf("error opening database: %v", err))
		}
		sharedDB = db
	})
	return sharedDB
}
