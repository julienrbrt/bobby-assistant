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

package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/pebble-dev/bobby-assistant/service/assistant/config"
)

func mapUnitsSystem(units string) string {
	switch units {
	case "imperial":
		return "IMPERIAL"
	default:
		return "METRIC"
	}
}

func mapTemperatureUnit(units string) string {
	switch units {
	case "imperial":
		return "°F"
	default:
		return "°C"
	}
}

func mapWindSpeedUnit(units string) string {
	switch units {
	case "metric":
		return "km/h"
	default:
		return "mph"
	}
}

var conditionToIconCode = map[string]int{
	"CLEAR":              8,
	"MOSTLY_CLEAR":       8,
	"PARTLY_CLOUDY":      7,
	"MOSTLY_CLOUDY":      5,
	"OVERCAST":           5,
	"LIGHT_RAIN":         1,
	"RAIN":               1,
	"HEAVY_RAIN":         2,
	"LIGHT_SNOW":         3,
	"SNOW":               3,
	"HEAVY_SNOW":         4,
	"THUNDERSTORM":       2,
	"HEAVY_THUNDERSTORM": 2,
	"FOG":                5,
	"HAZE":               6,
	"WINDY":              6,
	"DRIZZLE":            1,
	"FREEZING_RAIN":      1,
	"ICE_PELLETS":        3,
	"SLEET":              3,
	"BLOWING_SNOW":       4,
	"RAIN_SNOW":          3,
	"UNSPECIFIED":        6,
}

func iconCodeForCondition(condType string) int {
	if code, ok := conditionToIconCode[condType]; ok {
		return code
	}
	return 6
}

func intPtr(v int) *int             { return &v }
func strPtr(v string) *string       { return &v }
func float64Ptr(v float64) *float64 { return &v }
func roundInt(v float64) int        { return int(math.Round(v)) }

type googleCurrentConditions struct {
	WeatherCondition struct {
		Description struct {
			Text string `json:"text"`
		} `json:"description"`
		Type string `json:"type"`
	} `json:"weatherCondition"`
	Temperature struct {
		Degrees float64 `json:"degrees"`
		Unit    string  `json:"unit"`
	} `json:"temperature"`
	FeelsLikeTemperature struct {
		Degrees float64 `json:"degrees"`
	} `json:"feelsLikeTemperature"`
	RelativeHumidity int `json:"relativeHumidity"`
	UvIndex          int `json:"uvIndex"`
	Wind             struct {
		Speed struct {
			Value float64 `json:"value"`
		} `json:"speed"`
		Direction struct {
			Degrees  float64 `json:"degrees"`
			Cardinal string  `json:"cardinal"`
		} `json:"direction"`
		Gust struct {
			Value float64 `json:"value"`
		} `json:"gust"`
	} `json:"wind"`
	Visibility struct {
		Distance float64 `json:"distance"`
	} `json:"visibility"`
	CloudCover int `json:"cloudCover"`
	Precipitation struct {
		Probability struct {
			Percent int `json:"percent"`
		} `json:"probability"`
	} `json:"precipitation"`
	CurrentConditionsHistory *struct {
		MaxTemperature *struct {
			Degrees float64 `json:"degrees"`
		} `json:"maxTemperature"`
		MinTemperature *struct {
			Degrees float64 `json:"degrees"`
		} `json:"minTemperature"`
	} `json:"currentConditionsHistory"`
}

type googleDayPartForecast struct {
	WeatherCondition struct {
		Description struct {
			Text string `json:"text"`
		} `json:"description"`
		Type string `json:"type"`
	} `json:"weatherCondition"`
	RelativeHumidity int `json:"relativeHumidity"`
	UvIndex          int `json:"uvIndex"`
	Precipitation    struct {
		Probability struct {
			Percent int    `json:"percent"`
			Type    string `json:"type"`
		} `json:"probability"`
		Qpf struct {
			Quantity float64 `json:"quantity"`
		} `json:"qpf"`
	} `json:"precipitation"`
	Wind struct {
		Speed struct {
			Value float64 `json:"value"`
		} `json:"speed"`
		Direction struct {
			Cardinal string `json:"cardinal"`
		} `json:"direction"`
	} `json:"wind"`
	CloudCover int `json:"cloudCover"`
}

type googleDailyForecast struct {
	ForecastDays []struct {
		DisplayDate struct {
			Year  int `json:"year"`
			Month int `json:"month"`
			Day   int `json:"day"`
		} `json:"displayDate"`
		DaytimeForecast   *googleDayPartForecast `json:"daytimeForecast"`
		NighttimeForecast *googleDayPartForecast `json:"nighttimeForecast"`
		MaxTemperature    struct {
			Degrees float64 `json:"degrees"`
		} `json:"maxTemperature"`
		MinTemperature struct {
			Degrees float64 `json:"degrees"`
		} `json:"minTemperature"`
		SunEvents struct {
			SunriseTime string `json:"sunriseTime"`
			SunsetTime  string `json:"sunsetTime"`
		} `json:"sunEvents"`
		MoonEvents struct {
			MoonPhase     string   `json:"moonPhase"`
			MoonriseTimes []string `json:"moonriseTimes"`
			MoonsetTimes  []string `json:"moonsetTimes"`
		} `json:"moonEvents"`
	} `json:"forecastDays"`
}

type googleHourlyForecast struct {
	ForecastHours []struct {
		Interval struct {
			StartTime string `json:"startTime"`
		} `json:"interval"`
		WeatherCondition struct {
			Description struct {
				Text string `json:"text"`
			} `json:"description"`
			Type string `json:"type"`
		} `json:"weatherCondition"`
		Temperature struct {
			Degrees float64 `json:"degrees"`
		} `json:"temperature"`
		FeelsLikeTemperature struct {
			Degrees float64 `json:"degrees"`
		} `json:"feelsLikeTemperature"`
		RelativeHumidity int `json:"relativeHumidity"`
		UvIndex          int `json:"uvIndex"`
		Precipitation    struct {
			Probability struct {
				Percent int    `json:"percent"`
				Type    string `json:"type"`
			} `json:"probability"`
		} `json:"precipitation"`
		Wind struct {
			Speed struct {
				Value float64 `json:"value"`
			} `json:"speed"`
			Direction struct {
				Cardinal string `json:"cardinal"`
			} `json:"direction"`
		} `json:"wind"`
		CloudCover int `json:"cloudCover"`
	} `json:"forecastHours"`
}

func formatTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Format("2006-01-02T15:04:05-0700")
}

func doJSONRequest(ctx context.Context, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}
	return nil
}

func populateDayPartEntry(dp *ForecastDayPart, idx int, fc *googleDayPartForecast, name string, dayOrNight string) {
	dp.CloudCover[idx] = intPtr(fc.CloudCover)
	dp.DayOrNight[idx] = strPtr(dayOrNight)
	dp.DaypartName[idx] = strPtr(name)
	icon := iconCodeForCondition(fc.WeatherCondition.Type)
	dp.IconCode[idx] = intPtr(icon)
	dp.IconCodeExtend[idx] = intPtr(icon)
	dp.Narrative[idx] = strPtr(fc.WeatherCondition.Description.Text)
	dp.PrecipChance[idx] = intPtr(fc.Precipitation.Probability.Percent)
	if fc.Precipitation.Probability.Type != "" {
		dp.PrecipType[idx] = strPtr(fc.Precipitation.Probability.Type)
	}
	dp.Qpf[idx] = float64Ptr(fc.Precipitation.Qpf.Quantity)
	dp.RelativeHumidity[idx] = intPtr(fc.RelativeHumidity)
	dp.UvIndex[idx] = intPtr(fc.UvIndex)
	dp.WindDirectionCardinal[idx] = strPtr(fc.Wind.Direction.Cardinal)
	dp.WindSpeed[idx] = intPtr(roundInt(fc.Wind.Speed.Value))
	dp.WxPhraseLong[idx] = strPtr(fc.WeatherCondition.Description.Text)
	dp.WxPhraseShort[idx] = strPtr(fc.WeatherCondition.Description.Text)
}

func GetDailyForecast(ctx context.Context, lat, lon float64, units string) (*Forecast, error) {
	us := mapUnitsSystem(units)
	url := fmt.Sprintf(
		"https://weather.googleapis.com/v1/forecast/days:lookup?key=%s&location.latitude=%f&location.longitude=%f&days=7&unitsSystem=%s",
		config.GetConfig().GoogleMapsStaticKey, lat, lon, us,
	)

	var gResp googleDailyForecast
	if err := doJSONRequest(ctx, url, &gResp); err != nil {
		return nil, err
	}

	n := len(gResp.ForecastDays)
	dpSize := n * 2

	dp := &ForecastDayPart{
		CloudCover:            make([]*int, dpSize),
		DayOrNight:            make([]*string, dpSize),
		DaypartName:           make([]*string, dpSize),
		IconCode:              make([]*int, dpSize),
		IconCodeExtend:        make([]*int, dpSize),
		Narrative:             make([]*string, dpSize),
		PrecipChance:          make([]*int, dpSize),
		PrecipType:            make([]*string, dpSize),
		Qpf:                   make([]*float64, dpSize),
		QpfSnow:               make([]*float64, dpSize),
		RelativeHumidity:      make([]*int, dpSize),
		Temperature:           make([]*int, dpSize),
		TemperatureHeatIndex:  make([]*int, dpSize),
		TemperatureWindChill:  make([]*int, dpSize),
		UvIndex:               make([]*int, dpSize),
		WindDirectionCardinal: make([]*string, dpSize),
		WindSpeed:             make([]*int, dpSize),
		WxPhraseLong:          make([]*string, dpSize),
		WxPhraseShort:         make([]*string, dpSize),
	}

	forecast := &Forecast{
		CalendarDayTemperatureMax: make([]int, n),
		CalendarDayTemperatureMin: make([]int, n),
		DayOfWeek:                 make([]string, n),
		MoonPhaseCode:             make([]string, n),
		MoonPhase:                 make([]string, n),
		MoonPhaseDay:              make([]int, n),
		Narrative:                 make([]string, n),
		SunriseTimeLocal:          make([]string, n),
		SunsetTimeLocal:           make([]string, n),
		MoonriseTimeLocal:         make([]string, n),
		MoonsetTimeLocal:          make([]string, n),
		Qpf:                       make([]float32, n),
		QpfSnow:                   make([]float32, n),
		DayParts:                  []ForecastDayPart{*dp},
		WindSpeedUnit:             mapWindSpeedUnit(units),
		TemperatureUnit:           mapTemperatureUnit(units),
	}

	for i, day := range gResp.ForecastDays {
		date := time.Date(day.DisplayDate.Year, time.Month(day.DisplayDate.Month), day.DisplayDate.Day, 0, 0, 0, 0, time.Local)
		dayName := date.Weekday().String()

		forecast.DayOfWeek[i] = dayName
		forecast.CalendarDayTemperatureMax[i] = roundInt(day.MaxTemperature.Degrees)
		forecast.CalendarDayTemperatureMin[i] = roundInt(day.MinTemperature.Degrees)
		forecast.MoonPhase[i] = day.MoonEvents.MoonPhase
		forecast.MoonPhaseCode[i] = day.MoonEvents.MoonPhase

		if day.SunEvents.SunriseTime != "" {
			forecast.SunriseTimeLocal[i] = formatTimestamp(day.SunEvents.SunriseTime)
		}
		if day.SunEvents.SunsetTime != "" {
			forecast.SunsetTimeLocal[i] = formatTimestamp(day.SunEvents.SunsetTime)
		}

		if len(day.MoonEvents.MoonriseTimes) > 0 {
			forecast.MoonriseTimeLocal[i] = formatTimestamp(day.MoonEvents.MoonriseTimes[0])
		}
		if len(day.MoonEvents.MoonsetTimes) > 0 {
			forecast.MoonsetTimeLocal[i] = formatTimestamp(day.MoonEvents.MoonsetTimes[0])
		}

		var totalQpf float64
		if day.DaytimeForecast != nil {
			dayPartName := dayName
			if i == 0 {
				dayPartName = "Today"
			}
			populateDayPartEntry(dp, i*2, day.DaytimeForecast, dayPartName, "D")
			forecast.Narrative[i] = day.DaytimeForecast.WeatherCondition.Description.Text
			totalQpf += day.DaytimeForecast.Precipitation.Qpf.Quantity
		}
		if day.NighttimeForecast != nil {
			nightName := dayName + " Night"
			if i == 0 {
				nightName = "Tonight"
			}
			populateDayPartEntry(dp, i*2+1, day.NighttimeForecast, nightName, "N")
			totalQpf += day.NighttimeForecast.Precipitation.Qpf.Quantity
			if day.DaytimeForecast == nil {
				forecast.Narrative[i] = day.NighttimeForecast.WeatherCondition.Description.Text
			}
		}
		forecast.Qpf[i] = float32(totalQpf)
	}

	return forecast, nil
}

func GetCurrentConditions(ctx context.Context, lat, lon float64, units string) (*CurrentConditions, error) {
	us := mapUnitsSystem(units)
	url := fmt.Sprintf(
		"https://weather.googleapis.com/v1/currentConditions:lookup?key=%s&location.latitude=%f&location.longitude=%f&unitsSystem=%s",
		config.GetConfig().GoogleMapsStaticKey, lat, lon, us,
	)

	var gResp googleCurrentConditions
	if err := doJSONRequest(ctx, url, &gResp); err != nil {
		return nil, err
	}

	now := time.Now()
	conditions := &CurrentConditions{
		CloudCover:            gResp.CloudCover,
		DayOfWeek:             now.Weekday().String(),
		DayOrNight: func() string {
			if now.Hour() >= 6 && now.Hour() < 18 {
				return "D"
			}
			return "N"
		}(),
		RelativeHumidity:      gResp.RelativeHumidity,
		Temperature:           roundInt(gResp.Temperature.Degrees),
		TemperatureFeelsLike:  roundInt(gResp.FeelsLikeTemperature.Degrees),
		TemperatureUnit:       mapTemperatureUnit(units),
		UVIndex:               gResp.UvIndex,
		Visibility:            float32(gResp.Visibility.Distance),
		WindDirectionCardinal: gResp.Wind.Direction.Cardinal,
		WindSpeed:             roundInt(gResp.Wind.Speed.Value),
		WindSpeedUnit:         mapWindSpeedUnit(units),
		WindGust:              roundInt(gResp.Wind.Gust.Value),
		Description:           gResp.WeatherCondition.Description.Text,
		IconCode:              iconCodeForCondition(gResp.WeatherCondition.Type),
		Attribution:           "Google Weather",
	}

	if gResp.CurrentConditionsHistory != nil {
		if gResp.CurrentConditionsHistory.MaxTemperature != nil {
			conditions.TemperatureMax24Hour = roundInt(gResp.CurrentConditionsHistory.MaxTemperature.Degrees)
		}
		if gResp.CurrentConditionsHistory.MinTemperature != nil {
			conditions.TemperatureMin24Hour = roundInt(gResp.CurrentConditionsHistory.MinTemperature.Degrees)
		}
	}

	return conditions, nil
}

func GetHourlyForecast(ctx context.Context, lat, lon float64, units string) (*HourlyForecast, error) {
	us := mapUnitsSystem(units)
	url := fmt.Sprintf(
		"https://weather.googleapis.com/v1/forecast/hours:lookup?key=%s&location.latitude=%f&location.longitude=%f&hours=24&unitsSystem=%s",
		config.GetConfig().GoogleMapsStaticKey, lat, lon, us,
	)

	var gResp googleHourlyForecast
	if err := doJSONRequest(ctx, url, &gResp); err != nil {
		return nil, err
	}

	n := len(gResp.ForecastHours)
	hourly := &HourlyForecast{
		WxPhraseLong:          make([]string, n),
		Temperature:           make([]int, n),
		PrecipChance:          make([]int, n),
		PrecipType:            make([]string, n),
		ValidTimeLocal:        make([]string, n),
		UVIndex:               make([]int, n),
		WindSpeed:             make([]int, n),
		WindDirectionCardinal: make([]string, n),
		TemperatureUnit:       mapTemperatureUnit(units),
		WindSpeedUnit:         mapWindSpeedUnit(units),
	}

	for i, hour := range gResp.ForecastHours {
		hourly.WxPhraseLong[i] = hour.WeatherCondition.Description.Text
		hourly.Temperature[i] = roundInt(hour.Temperature.Degrees)
		hourly.PrecipChance[i] = hour.Precipitation.Probability.Percent
		hourly.PrecipType[i] = hour.Precipitation.Probability.Type
		hourly.ValidTimeLocal[i] = formatTimestamp(hour.Interval.StartTime)
		hourly.UVIndex[i] = hour.UvIndex
		hourly.WindSpeed[i] = roundInt(hour.Wind.Speed.Value)
		hourly.WindDirectionCardinal[i] = hour.Wind.Direction.Cardinal
	}

	return hourly, nil
}

type Forecast struct {
	CalendarDayTemperatureMax []int             `json:"calendarDayTemperatureMax"`
	CalendarDayTemperatureMin []int             `json:"calendarDayTemperatureMin"`
	DayOfWeek                 []string          `json:"dayOfWeek"`
	MoonPhaseCode             []string          `json:"moonPhaseCode"`
	MoonPhase                 []string          `json:"moonPhase"`
	MoonPhaseDay              []int             `json:"moonPhaseDay"`
	Narrative                 []string          `json:"narrative"`
	SunriseTimeLocal          []string          `json:"sunriseTimeLocal"`
	SunsetTimeLocal           []string          `json:"sunsetTimeLocal"`
	MoonriseTimeLocal         []string          `json:"moonriseTimeLocal"`
	MoonsetTimeLocal          []string          `json:"moonsetTimeLocal"`
	Qpf                       []float32         `json:"qpf"`
	QpfSnow                   []float32         `json:"qpfSnow"`
	DayParts                  []ForecastDayPart `json:"daypart"`
	WindSpeedUnit             string            `json:"windSpeedUnit"`
	TemperatureUnit           string            `json:"temperatureUnit"`
}

type ForecastDayPart struct {
	CloudCover            []*int     `json:"cloudCover"`
	DayOrNight            []*string  `json:"dayOrNight"`
	DaypartName           []*string  `json:"daypartName"`
	IconCode              []*int     `json:"iconCode"`
	IconCodeExtend        []*int     `json:"iconCodeExtend"`
	Narrative             []*string  `json:"narrative"`
	PrecipChance          []*int     `json:"precipChance"`
	PrecipType            []*string  `json:"precipType"`
	Qpf                   []*float64 `json:"qpf"`
	QpfSnow               []*float64 `json:"qpfSnow"`
	QualifierCode         []*string  `json:"qualifierCode"`
	QualifierPhrase       []*string  `json:"qualifierPhrase"`
	RelativeHumidity      []*int     `json:"relativeHumidity"`
	SnowRange             []*string  `json:"snowRange"`
	Temperature           []*int     `json:"temperature"`
	TemperatureHeatIndex  []*int     `json:"temperatureHeatIndex"`
	TemperatureWindChill  []*int     `json:"temperatureWindChill"`
	ThunderCategory       []*string  `json:"thunderCategory"`
	ThunderIndex          []*int     `json:"thunderIndex"`
	UvDescription         []*string  `json:"uvDescription"`
	UvIndex               []*int     `json:"uvIndex"`
	WindDirection         []*int     `json:"windDirection"`
	WindDirectionCardinal []*string  `json:"windDirectionCardinal"`
	WindPhrase            []*string  `json:"windPhrase"`
	WindSpeed             []*int     `json:"windSpeed"`
	WxPhraseLong          []*string  `json:"wxPhraseLong"`
	WxPhraseShort         []*string  `json:"wxPhraseShort"`
}

type CurrentConditions struct {
	CloudCoverPhrase      string  `json:"cloudCoverPhrase"`
	CloudCover            int     `json:"cloudCover"`
	DayOfWeek             string  `json:"dayOfWeek"`
	DayOrNight            string  `json:"dayOrNight"`
	Precip1Hour           float32 `json:"precip1Hour"`
	Precip6Hour           float32 `json:"precip6Hour"`
	Precip12Hour          float32 `json:"precip12Hour"`
	RelativeHumidity      int     `json:"relativeHumidity"`
	SunriseTimeLocal      string  `json:"sunriseTimeLocal"`
	SunsetTimeLocal       string  `json:"sunsetTimeLocal"`
	Temperature           int     `json:"temperature"`
	TemperatureFeelsLike  int     `json:"temperatureFeelsLike"`
	TemperatureMax24Hour  int     `json:"temperatureMax24Hour"`
	TemperatureMin24Hour  int     `json:"temperatureMin24Hour"`
	TemperatureWindChill  int     `json:"temperatureWindChill"`
	TemperatureUnit       string  `json:"temperatureUnit"`
	UVIndex               int     `json:"uvIndex"`
	Visibility            float32 `json:"visibility"`
	WindDirectionCardinal string  `json:"windDirectionCardinal"`
	WindSpeed             int     `json:"windSpeed"`
	WindSpeedUnit         string  `json:"windSpeedUnit"`
	WindGust              int     `json:"windGust"`
	Description           string  `json:"wxPhraseLong"`
	IconCode              int     `json:"iconCode"`
	Attribution           string  `json:"attribution"`
}

type HourlyForecast struct {
	WxPhraseLong          []string `json:"wxPhraseLong"`
	Temperature           []int    `json:"temperature"`
	PrecipChance          []int    `json:"precipChance"`
	PrecipType            []string `json:"precipType"`
	ValidTimeLocal        []string `json:"validTimeLocal"`
	UVIndex               []int    `json:"uvIndex"`
	WindSpeed             []int    `json:"windSpeed"`
	WindDirectionCardinal []string `json:"windDirectionCardinal"`
	TemperatureUnit       string   `json:"temperatureUnit"`
	WindSpeedUnit         string   `json:"windSpeedUnit"`
}
