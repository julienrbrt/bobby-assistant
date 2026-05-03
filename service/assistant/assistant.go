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

package assistant

import (
	"log"
	"net/http"

	"gorm.io/gorm"
)

type Service struct {
	mux *http.ServeMux
	db  *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	s := &Service{
		mux: http.NewServeMux(),
		db:  db,
	}
	s.mux.HandleFunc("/query", s.handleQuery)
	s.mux.HandleFunc("/heartbeat", s.handleHeartbeat)
	s.mux.HandleFunc("/robots.txt", s.handleRobots)
	return s
}

func (s *Service) handleHeartbeat(rw http.ResponseWriter, r *http.Request) {
	_, _ = rw.Write([]byte("bobby"))
}

func (s *Service) handleQuery(rw http.ResponseWriter, r *http.Request) {
	session, err := NewPromptSession(s.db, rw, r)
	if err != nil {
		log.Printf("Creating session failed: %v", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Run(r.Context())
}

func (s *Service) handleRobots(rw http.ResponseWriter, r *http.Request) {
	_, _ = rw.Write([]byte("User-agent: *\nDisallow: /\n"))
}

func (s *Service) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(rw, r)
}

func (s *Service) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}
