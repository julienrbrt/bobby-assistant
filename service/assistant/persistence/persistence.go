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

package persistence

import (
	"context"
	"encoding/json"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/pebble-dev/bobby-assistant/service/assistant/llm"
	"github.com/pebble-dev/bobby-assistant/service/assistant/util"
)

type SerializedMessage struct {
	Role             string                `json:"role"`
	Content          string                `json:"content"`
	FunctionCall     *llm.FunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *llm.FunctionResponse `json:"functionResponse,omitempty"`
}

type StoredContext struct {
	PoiQuery  *util.POIQuery `json:"poiQuery"`
	POIs      []util.POI     `json:"pois"`
	LastRoute map[string]any `json:"lastRoute"`
}

type ThreadContext struct {
	ThreadId       uuid.UUID           `json:"threadId"`
	Messages       []SerializedMessage `json:"messages"`
	ContextStorage StoredContext       `json:"contextStorage"`
}

type Thread struct {
	ID   string `gorm:"primaryKey"`
	Data string
}

func NewContext() *ThreadContext {
	return &ThreadContext{}
}

func InitDB(db *gorm.DB) {
	db.AutoMigrate(&Thread{})
}

func LoadThread(ctx context.Context, db *gorm.DB, id string) (*ThreadContext, error) {
	span := sentry.StartSpan(ctx, "load_thread")
	ctx = span.Context()
	defer span.Finish()
	var thread Thread
	if err := db.WithContext(ctx).Where("id = ?", id).First(&thread).Error; err != nil {
		return nil, err
	}
	var threadContext ThreadContext
	if err := json.Unmarshal([]byte(thread.Data), &threadContext); err != nil {
		return nil, err
	}
	return &threadContext, nil
}

func StoreThread(ctx context.Context, db *gorm.DB, thread *ThreadContext) error {
	span := sentry.StartSpan(ctx, "store_thread")
	ctx = span.Context()
	defer span.Finish()
	j, err := json.Marshal(thread)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Save(&Thread{
		ID:   thread.ThreadId.String(),
		Data: string(j),
	}).Error
}
