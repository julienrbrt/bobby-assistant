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
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

type TimeResponse struct {
	Time string `json:"time"`
}

type GetTimeInput struct {
	// The timezone, e.g. 'America/Los_Angeles'.
	Timezone string `json:"timezone" jsonschema:"required"`
	// The number of seconds to add to the current time.
	Offset float64 `json:"offset"`
}

func init() {
	registerFunction(Registration{
		Definition: shared.FunctionDefinitionParam{
			Name:        "get_time_elsewhere",
			Description: openai.String("Get the current time in a given valid tzdb timezone. Not all cities have a tzdb entry - be sure to use one that exists. Call multiple times to find the time in multiple timezones."),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"timezone": map[string]any{
						"type":        "string",
						"description": "The timezone, e.g. 'America/Los_Angeles'.",
					},
					"offset": map[string]any{
						"type":        "number",
						"description": "The number of seconds to add to the current time, if checking a different time. Omit or set to zero for current time.",
						"format":      "double",
					},
				},
				"required": []string{"timezone"},
			},
		},
		Fn:        getTimeElsewhere,
		Thought:   getTimeThought,
		InputType: GetTimeInput{},
	})
}

func getTimeThought(args any) string {
	arg := args.(*GetTimeInput)
	if arg.Timezone != "" {
		s := strings.Split(arg.Timezone, "/")
		place := strings.Replace(s[len(s)-1], "_", " ", -1)
		return "Checking the time in " + place
	}
	return "Checking the time"
}

func getTimeElsewhere(ctx context.Context, args any) any {
	span := sentry.StartSpan(ctx, "get_time_elsewhere")
	defer span.Finish()
	arg := args.(*GetTimeInput)
	utc := time.Now().UTC().Add(time.Duration(arg.Offset) * time.Second)
	loc, err := time.LoadLocation(arg.Timezone)
	if err != nil {
		return Error{fmt.Sprintf("The timezone %q is not valid", arg.Timezone)}
	}
	utc.In(loc)
	return TimeResponse{utc.In(loc).Format(time.RFC1123)}
}
