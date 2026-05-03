package geocoding

import (
	"context"
	"fmt"

	"github.com/getsentry/sentry-go"
	gmaps "googlemaps.github.io/maps"
	"github.com/pebble-dev/bobby-assistant/service/assistant/query"
)

type Location struct {
	Lat float64
	Lon float64
}

var client *gmaps.Client

func Init(apiKey, signingSecret string) error {
	var err error
	if signingSecret != "" {
		client, err = gmaps.NewClient(gmaps.WithAPIKeyAndSignature(apiKey, signingSecret))
	} else {
		client, err = gmaps.NewClient(gmaps.WithAPIKey(apiKey))
	}
	return err
}

func Geocode(ctx context.Context, search string) (Location, error) {
	span := sentry.StartSpan(ctx, "google.geocoding")
	ctx = span.Context()
	defer span.Finish()
	userLocation := query.LocationFromContext(ctx)
	req := &gmaps.GeocodingRequest{
		Address: search,
	}
	if userLocation != nil {
		req.Bounds = &gmaps.LatLngBounds{
			NorthEast: gmaps.LatLng{Lat: userLocation.Lat + 0.5, Lng: userLocation.Lon + 0.5},
			SouthWest: gmaps.LatLng{Lat: userLocation.Lat - 0.5, Lng: userLocation.Lon - 0.5},
		}
	}
	results, err := client.Geocode(ctx, req)
	if err != nil {
		return Location{}, fmt.Errorf("could not find location: %w", err)
	}
	if len(results) == 0 {
		return Location{}, fmt.Errorf("could not find location with name %q", search)
	}
	return Location{
		Lat: results[0].Geometry.Location.Lat,
		Lon: results[0].Geometry.Location.Lng,
	}, nil
}

func ReverseGeocode(ctx context.Context, lon, lat float64) (*gmaps.GeocodingResult, error) {
	span := sentry.StartSpan(ctx, "google.reverse_geocoding")
	ctx = span.Context()
	defer span.Finish()
	req := &gmaps.GeocodingRequest{
		LatLng:   &gmaps.LatLng{Lat: lat, Lng: lon},
		ResultType: []string{"locality", "administrative_area_level_1", "country"},
	}
	results, err := client.ReverseGeocode(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not reverse geocode location: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("the user isn't anywhere")
	}
	return &results[0], nil
}
