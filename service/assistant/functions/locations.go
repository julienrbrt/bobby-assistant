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

package functions

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	"github.com/pebble-dev/bobby-assistant/service/assistant/query"
	"github.com/pebble-dev/bobby-assistant/service/assistant/util/mapbox"
	"github.com/umahmood/haversine"
)

type LocationResponse struct {
	Latitude           float64 `json:"latitude"`
	Longitude          float64 `json:"longitude"`
	DistanceKilometers float64 `json:"distance_meters,omitempty"`
	DistanceMiles      float64 `json:"distance_miles,omitempty"`
}

type GetLocationInput struct {
	// The name of a place to locate, e.g. "San Francisco, CA, USA" or "Paris, France".
	PlaceName string `json:"place_name"`
}

func init() {
	registerFunction(Registration{
		Definition: shared.FunctionDefinitionParam{
			Name:        "get_location",
			Description: openai.String("Get the latitude and longitude of a given location. If the user's location is available, also provides the distance from the user."),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"place_name": map[string]any{
						"type":        "string",
						"description": `The name of a place to locate, e.g. "San Francisco, CA, USA" or "Paris, France".`,
					},
				},
				"required": []string{"place_name"},
			},
		},
		Fn:        getLocationImpl,
		Thought:   getLocationThought,
		InputType: GetLocationInput{},
	})
}

func getLocationThought(args any) string {
	arg := args.(*GetLocationInput)
	return fmt.Sprintf("Locating %q", arg.PlaceName)
}

func getLocationImpl(ctx context.Context, args any) any {
	span := sentry.StartSpan(ctx, "get_location")
	ctx = span.Context()
	defer span.Finish()
	arg := args.(*GetLocationInput)
	location, err := mapbox.GeocodeWithContext(ctx, arg.PlaceName)
	if err != nil {
		return fmt.Errorf("failed to geocode %q: %w", arg.PlaceName, err)
	}
	userLocation := query.LocationFromContext(ctx)
	lr := LocationResponse{
		Latitude:  location.Lat,
		Longitude: location.Lon,
	}
	if userLocation != nil {
		uh := haversine.Coord{Lat: userLocation.Lat, Lon: userLocation.Lon}
		lh := haversine.Coord{Lat: location.Lat, Lon: location.Lon}
		lr.DistanceMiles, lr.DistanceKilometers = haversine.Distance(uh, lh)
	}
	return lr
}
