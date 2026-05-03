// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
	"github.com/glebarez/sqlite"

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
