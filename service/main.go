package main

import (
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/pebble-dev/bobby-assistant/service/assistant"
	"github.com/pebble-dev/bobby-assistant/service/assistant/config"
	"github.com/pebble-dev/bobby-assistant/service/assistant/persistence"
	"github.com/pebble-dev/bobby-assistant/service/assistant/util/geocoding"
	"github.com/pebble-dev/bobby-assistant/service/assistant/util/redact"
	"github.com/pebble-dev/bobby-assistant/service/assistant/util/storage"
)

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.GetConfig().SentryDSN,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		BeforeSend:       redact.BeforeSend,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)

	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	db := storage.GetDB()
	persistence.InitDB(db)
	cfg := config.GetConfig()
	if cfg.GoogleMapsStaticKey != "" {
		if err := geocoding.Init(cfg.GoogleMapsStaticKey, cfg.GoogleMapsStaticSecret); err != nil {
			log.Fatalf("geocoding.Init: %s", err)
		}
	}
	service := assistant.NewService(db)
	addr := "0.0.0.0:8080"
	log.Printf("Listening on %s.", addr)
	log.Fatal(http.ListenAndServe(addr, sentryHandler.Handle(service)))
}
