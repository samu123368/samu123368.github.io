package accuweather

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const apiURL = "https://api.open-meteo.com/v1/forecast"

var httpClient = &http.Client{Timeout: 30 * time.Second}

type openMeteoResponse struct {
	UTCOffsetSeconds int `json:"utc_offset_seconds"`
	Current          struct {
		Time        string  `json:"time"`
		Temperature float64 `json:"temperature_2m"`
		WeatherCode int     `json:"weather_code"`
		IsDay       int     `json:"is_day"`
	} `json:"current"`
	Hourly struct {
		WeatherCode []int `json:"weather_code"`
	} `json:"hourly"`
	Daily struct {
		WeatherCode              []int     `json:"weather_code"`
		TemperatureMax           []float64 `json:"temperature_2m_max"`
		TemperatureMin           []float64 `json:"temperature_2m_min"`
		PrecipitationProbability []float64 `json:"precipitation_probability_max"`
		WindSpeedMax             []float64 `json:"wind_speed_10m_max"`
		WindDirectionDominant    []float64 `json:"wind_direction_10m_dominant"`
	} `json:"daily"`
}

type Coordinate struct {
	Longitude float64
	Latitude  float64
}

func GetWeather(longitude float64, latitude float64, currentTime int64, _ string) *Weather {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		weather, err := getOpenMeteoWeather(longitude, latitude, currentTime)
		if err == nil {
			return weather
		}
		lastErr = err
		time.Sleep(time.Duration(attempt+1) * 5 * time.Second)
	}

	fmt.Printf("Open-Meteo request failed for %.6f,%.6f: %v; using blank data\n", latitude, longitude, lastErr)
	return BlankData()
}

func getOpenMeteoWeather(longitude float64, latitude float64, currentTime int64) (*Weather, error) {
	weather, err := getOpenMeteoWeatherBatch([]Coordinate{{Longitude: longitude, Latitude: latitude}}, currentTime)
	if err != nil {
		return nil, err
	}
	return weather[0], nil
}

// GetWeatherBatch requests multiple locations in one Open-Meteo call. The
// returned slice always corresponds position-for-position with coordinates.
func GetWeatherBatch(coordinates []Coordinate, currentTime int64, _ string) ([]*Weather, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		weather, err := getOpenMeteoWeatherBatch(coordinates, currentTime)
		if err == nil {
			return weather, nil
		}
		lastErr = err
		time.Sleep(time.Duration(attempt+1) * 5 * time.Second)
	}
	return nil, fmt.Errorf("Open-Meteo batch failed after retries: %w", lastErr)
}

func getOpenMeteoWeatherBatch(coordinates []Coordinate, currentTime int64) ([]*Weather, error) {
	if len(coordinates) == 0 {
		return nil, nil
	}

	latitudes := make([]string, len(coordinates))
	longitudes := make([]string, len(coordinates))
	for i, coordinate := range coordinates {
		latitudes[i] = strconv.FormatFloat(coordinate.Latitude, 'f', 6, 64)
		longitudes[i] = strconv.FormatFloat(coordinate.Longitude, 'f', 6, 64)
	}

	query := url.Values{}
	query.Set("latitude", strings.Join(latitudes, ","))
	query.Set("longitude", strings.Join(longitudes, ","))
	query.Set("current", "temperature_2m,weather_code,is_day")
	query.Set("hourly", "weather_code")
	query.Set("daily", "weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max,wind_speed_10m_max,wind_direction_10m_dominant")
	query.Set("temperature_unit", "celsius")
	query.Set("wind_speed_unit", "kmh")
	query.Set("timezone", "auto")
	query.Set("forecast_days", "8")

	response, err := httpClient.Get(apiURL + "?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %s", response.Status)
	}

	var raw json.RawMessage
	if err := json.NewDecoder(response.Body).Decode(&raw); err != nil {
		return nil, err
	}

	var data []openMeteoResponse
	if len(raw) > 0 && raw[0] == '[' {
		if err := json.Unmarshal(raw, &data); err != nil {
			return nil, err
		}
	} else {
		var single openMeteoResponse
		if err := json.Unmarshal(raw, &single); err != nil {
			return nil, err
		}
		data = []openMeteoResponse{single}
	}
	if len(data) != len(coordinates) {
		return nil, fmt.Errorf("forecast response count mismatch: got %d, want %d", len(data), len(coordinates))
	}

	weather := make([]*Weather, len(data))
	for i := range data {
		parsed, err := weatherFromOpenMeteo(data[i], currentTime)
		if err != nil {
			return nil, fmt.Errorf("location %d: %w", i, err)
		}
		weather[i] = parsed
	}
	return weather, nil
}

func weatherFromOpenMeteo(data openMeteoResponse, currentTime int64) (*Weather, error) {
	if len(data.Daily.WeatherCode) < 8 || len(data.Daily.TemperatureMax) < 8 || len(data.Daily.TemperatureMin) < 8 || len(data.Daily.PrecipitationProbability) < 8 || len(data.Daily.WindSpeedMax) < 2 || len(data.Daily.WindDirectionDominant) < 2 || len(data.Hourly.WeatherCode) < 43 {
		return nil, fmt.Errorf("incomplete forecast response")
	}

	weather := BlankData()
	weather.LocalTime = data.Current.Time
	weather.Current.TempCelsius = data.Current.Temperature
	weather.Current.TempFahrenheit = cToF(data.Current.Temperature)
	weather.Current.WeatherIcon = wmoToAccuWeather(data.Current.WeatherCode, data.Current.IsDay == 1)
	weather.Current.WindMetric = data.Daily.WindSpeedMax[0]
	weather.Current.WindImperial = kmToMPH(data.Daily.WindSpeedMax[0])
	weather.Current.WindDirection = degreesToDirection(data.Daily.WindDirectionDominant[0])
	weather.Globe.Offset = data.UTCOffsetSeconds / 3600
	weather.Globe.Time = int(currentTime) + data.UTCOffsetSeconds

	weather.Today.TempCelsiusMin = data.Daily.TemperatureMin[0]
	weather.Today.TempCelsiusMax = data.Daily.TemperatureMax[0]
	weather.Today.TempFahrenheitMin = cToF(data.Daily.TemperatureMin[0])
	weather.Today.TempFahrenheitMax = cToF(data.Daily.TemperatureMax[0])
	weather.Today.WeatherIcon = wmoToAccuWeather(data.Daily.WeatherCode[0], true)

	weather.Tomorrow.TempCelsiusMin = data.Daily.TemperatureMin[1]
	weather.Tomorrow.TempCelsiusMax = data.Daily.TemperatureMax[1]
	weather.Tomorrow.TempFahrenheitMin = cToF(data.Daily.TemperatureMin[1])
	weather.Tomorrow.TempFahrenheitMax = cToF(data.Daily.TemperatureMax[1])
	weather.Tomorrow.WeatherIcon = wmoToAccuWeather(data.Daily.WeatherCode[1], true)

	weather.Wind.WindMetric = data.Daily.WindSpeedMax[0]
	weather.Wind.WindImperial = kmToMPH(data.Daily.WindSpeedMax[0])
	weather.Wind.WindDirection = degreesToDirection(data.Daily.WindDirectionDominant[0])
	weather.Wind.WindMetricTomorrow = data.Daily.WindSpeedMax[1]
	weather.Wind.WindImperialTomorrow = kmToMPH(data.Daily.WindSpeedMax[1])
	weather.Wind.WindDirectionTomorrow = degreesToDirection(data.Daily.WindDirectionDominant[1])
	weather.UVIndex = 0
	weather.Pollen = 2

	hourIndexes := []int{0, 6, 12, 18, 24, 30, 36, 42}
	weather.HourlyIcon = make([]int, 8)
	weather.Precipitation = make([]int, 15)
	for i, hourIndex := range hourIndexes {
		localHour := hourIndex % 24
		isDay := localHour >= 6 && localHour < 20
		weather.HourlyIcon[i] = wmoToAccuWeather(data.Hourly.WeatherCode[hourIndex], isDay)
		weather.Precipitation[i] = 255
	}

	weather.Week = make([]Week, 7)
	for i := 0; i < 7; i++ {
		day := i + 1
		weather.Week[i] = Week{
			TempCelsiusMin:    data.Daily.TemperatureMin[day],
			TempCelsiusMax:    data.Daily.TemperatureMax[day],
			TempFahrenheitMin: cToF(data.Daily.TemperatureMin[day]),
			TempFahrenheitMax: cToF(data.Daily.TemperatureMax[day]),
			WeatherIcon:       wmoToAccuWeather(data.Daily.WeatherCode[day], true),
		}
		weather.Precipitation[8+i] = clampInt(int(math.Round(data.Daily.PrecipitationProbability[day])), 0, 100)
	}

	return weather, nil
}

func cToF(celsius float64) float64 {
	return celsius*9/5 + 32
}

func kmToMPH(km float64) float64 {
	return km / 1.60934
}

func degreesToDirection(degrees float64) string {
	directions := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	index := int(math.Floor((degrees+11.25)/22.5)) % len(directions)
	if index < 0 {
		index += len(directions)
	}
	return directions[index]
}

func clampInt(value int, minimum int, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func wmoToAccuWeather(code int, isDay bool) int {
	switch code {
	case 0:
		if isDay {
			return 1
		}
		return 33
	case 1:
		if isDay {
			return 2
		}
		return 34
	case 2:
		if isDay {
			return 3
		}
		return 35
	case 3:
		if isDay {
			return 7
		}
		return 38
	case 45, 48:
		return 11
	case 51, 61, 80:
		return 12
	case 53, 55, 63, 65, 81, 82:
		return 18
	case 56, 57, 66, 67:
		return 26
	case 71, 77:
		return 19
	case 73, 75, 86:
		return 22
	case 85:
		return 21
	case 95, 96, 99:
		return 15
	default:
		return 7
	}
}

func BlankData() *Weather {
	week := make([]Week, 7)
	for i := range week {
		week[i] = Week{
			TempFahrenheitMin: -128,
			TempFahrenheitMax: -128,
			TempCelsiusMin:    -128,
			TempCelsiusMax:    -128,
			WeatherIcon:       0,
		}
	}

	return &Weather{
		Current:       Current{TempFahrenheit: -128, TempCelsius: -128, WindDirection: "N"},
		Today:         Today{TempFahrenheitMin: -128, TempFahrenheitMax: -128, TempCelsiusMin: -128, TempCelsiusMax: -128},
		Tomorrow:      Tomorrow{TempFahrenheitMin: -128, TempFahrenheitMax: -128, TempCelsiusMin: -128, TempCelsiusMax: -128},
		Week:          week,
		Wind:          Wind{WindDirection: "N", WindDirectionTomorrow: "N"},
		Precipitation: []int{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		HourlyIcon:    []int{0, 0, 0, 0, 0, 0, 0, 0},
		Globe:         Globe{Time: int(time.Now().Unix())},
	}
}
