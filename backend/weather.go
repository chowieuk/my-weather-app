package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type WeatherApiResponse struct {
	Location Location            `json:"location"`
	Forecast map[string]Forecast `json:"forecast"`
}

type Location struct {
	Name      string `json:"name"`
	Country   string `json:"country"`
	Region    string `json:"region"`
	LocalTime string `json:"localtime"`
}

type Forecast struct {
	Date      string    `json:"date"`
	DateEpoch int64     `json:"date_epoch"`
	AstroData AstroData `json:"astro"`
	MinTemp   int       `json:"mintemp"`
	MaxTemp   int       `json:"maxtemp"`
	AvgTemp   int       `json:"avgtemp"`
	TotalSnow int       `json:"totalsnow"`
	SunHour   float64   `json:"sunhour"`
	UVIndex   int       `json:"uv_index"`
}

type AstroData struct {
	Sunrise          string `json:"sunrise"`
	Sunset           string `json:"sunset"`
	Moonrise         string `json:"moonrise"`
	Moonset          string `json:"moonset"`
	MoonPhase        string `json:"moon_phase"`
	MoonIllumination int    `json:"moon_illumination"`
}

type AstroResponse struct {
	Date      string    `json:"date"`
	AstroData AstroData `json:"astro"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ApiErrorResponse struct {
	Success bool     `json:"success"`
	Error   ApiError `json:"error"`
}

type ApiError struct {
	Code int    `json:"code"`
	Type string `json:"type"`
	Info string `json:"info"`
}

type TimeManager interface {
	Now() time.Time
	ComputeExpiryTime(localTime string) (time.Time, error)
}

type DefaultTimeManager struct{}

func (tm *DefaultTimeManager) Now() time.Time {
	return time.Now()
}

func (tm *DefaultTimeManager) ComputeExpiryTime(localTime string) (time.Time, error) {
	parsedTime, err := time.Parse("2006-01-02 15:04", localTime)
	if err != nil {
		return time.Time{}, err
	}

	nextDay := parsedTime.AddDate(0, 0, 1)
	expiryTime := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, parsedTime.Location())

	return expiryTime, nil
}

type WeatherManager interface {
	GetAstro(location string) (AstroResponse, error)
}

type WeatherStackManager struct {
	BaseUrl     string
	WEATHER_KEY string
	Client      http.Client
	Cache       map[string]AstroResponse
	TimeManager TimeManager
}

func (wsm *WeatherStackManager) GetAstro(location string) (AstroResponse, error) {

	// Check the cache first
	currentDate := wsm.TimeManager.Now().Format("2006-01-02")
	cacheKey := makeKey(location, currentDate)

	if cachedAstroResponse, found := wsm.Cache[cacheKey]; found {
		if wsm.TimeManager.Now().Before(cachedAstroResponse.ExpiresAt) {
			log.Println("value served from cache", cachedAstroResponse)
			return cachedAstroResponse, nil
		}
		log.Printf("cached value for %s has expired\n", location)
	}

	url := fmt.Sprintf("%s/forecast?access_key=%s&query=%s", wsm.BaseUrl, wsm.WEATHER_KEY, location)
	response, err := wsm.Client.Get(url)
	if err != nil {
		log.Println(err)
		return AstroResponse{}, fmt.Errorf("failed to send request to Weather API: %w", err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return AstroResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		var apiErrorResponse ApiErrorResponse
		err = json.Unmarshal(body, &apiErrorResponse)
		if err != nil {
			log.Println(err)
			return AstroResponse{}, err
		}
		log.Printf("%+v\n", &apiErrorResponse)
		return AstroResponse{}, fmt.Errorf("API error occurred: code=%d, type=%s, info=%s", apiErrorResponse.Error.Code, apiErrorResponse.Error.Type, apiErrorResponse.Error.Info)
	}

	var forecastResponse WeatherApiResponse
	err = json.Unmarshal(body, &forecastResponse)
	if err != nil {
		log.Println(err)
		return AstroResponse{}, fmt.Errorf("failed to unmarshal forecast response: %w", err)
	}

	expiryTime, err := wsm.TimeManager.ComputeExpiryTime(forecastResponse.Location.LocalTime)
	if err != nil {
		log.Println(err)
		return AstroResponse{}, fmt.Errorf("failed to compute expiry time: %w", err)
	}

	// Extract the date from the forecast map (should be only one key-value pair)
	var forecastDate string
	var forecast Forecast
	for date, f := range forecastResponse.Forecast {
		forecastDate = date
		forecast = f
		break
	}

	astroResponse := AstroResponse{
		AstroData: forecast.AstroData,
		Date:      forecastDate,
		ExpiresAt: expiryTime,
	}

	wsm.Cache[cacheKey] = astroResponse

	log.Println("value served from fresh API request", astroResponse)
	return astroResponse, nil
}

func makeKey(location string, date string) string {
	return fmt.Sprintf("%s:%s", location, date)
}

func AstroHandler(w http.ResponseWriter, r *http.Request, wm WeatherManager) {
	// Parse location and date from request's query parameters
	location := r.URL.Query().Get("location")

	if location == "" {
		http.Error(w, "Missing 'location' query parameter", http.StatusBadRequest)
		return
	}

	// Get the astro data
	astroData, err := wm.GetAstro(location)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to get astro data", http.StatusInternalServerError)
		return
	}

	// Send the astro data as response
	// Convert the astroData to JSON before sending it
	astroDataJson, err := json.Marshal(astroData)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to convert astro data to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(astroDataJson)
}

func main() {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the WEATHER_KEY from the environment variables
	WEATHER_KEY := os.Getenv("WEATHER_KEY")

	wsm := WeatherStackManager{
		BaseUrl:     "http://api.weatherstack.com",
		WEATHER_KEY: WEATHER_KEY,
		Cache:       make(map[string]AstroResponse),
	}

	// Wrap the AstroHandler in a function that matches the expected signature
	http.HandleFunc("/astro", func(w http.ResponseWriter, r *http.Request) {
		AstroHandler(w, r, &wsm)
	})
	fmt.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
