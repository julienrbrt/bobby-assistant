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

package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// TODO: something reasonable.

type Config struct {
	BaseURL                string
	DBPath                 string
	IBMKey                 string
	ExchangeRateApiKey     string
	SentryDSN              string
	GoogleMapsStaticKey    string
	GoogleMapsStaticSecret string
	GoogleMapsStaticMapId  string
}

var c Config

func GetConfig() *Config {
	return &c
}

func init() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Only log if the file exists but couldn't be loaded
		if !os.IsNotExist(err) {
			log.Printf("Error loading .env file: %v", err)
		}
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "bobby.db"
	}
	c = Config{
		BaseURL:                os.Getenv("BASE_URL"),
		DBPath:                 dbPath,
		IBMKey:                 os.Getenv("IBM_KEY"),
		ExchangeRateApiKey:     os.Getenv("EXCHANGE_RATE_API_KEY"),
		SentryDSN:              os.Getenv("SENTRY_DSN"),
		GoogleMapsStaticKey:    os.Getenv("GOOGLE_MAPS_STATIC_KEY"),
		GoogleMapsStaticSecret: os.Getenv("GOOGLE_MAPS_STATIC_SECRET"),
		GoogleMapsStaticMapId:  os.Getenv("GOOGLE_MAPS_STATIC_MAP_ID"),
	}
}
