package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

type WeatherApiResponse struct {
	Location Location            `json:"location"`
	Forecast map[string]Forecast `json:"forecast"`
}

type Location struct {
	Name       string `json:"name"`
	Country    string `json:"country"`
	Region     string `json:"region"`
	LocalTime  string `json:"localtime"`
	TimeZoneID string `json:"timezone_id"`
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
	Date            string    `json:"date"`
	LocationName    string    `json:"name"`
	LocationRegion  string    `json:"region"`
	LocationCountry string    `json:"country"`
	AstroData       AstroData `json:"astro"`
	ExpiresAt       time.Time `json:"expires_at"`
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
	ComputeExpiryTime(localTime string, timeZoneID string) (time.Time, error)
}

type DefaultTimeManager struct{}

func (tm *DefaultTimeManager) Now() time.Time {
	return time.Now()
}

func (tm *DefaultTimeManager) ComputeExpiryTime(localTime string, timeZoneID string) (time.Time, error) {
	// Get the location for the given time zone
	loc, err := time.LoadLocation(timeZoneID)
	if err != nil {
		return time.Time{}, err
	}

	// Parse the local time using the location
	parsedTime, err := time.ParseInLocation("2006-01-02 15:04", localTime, loc)
	if err != nil {
		return time.Time{}, err
	}

	// Compute the expiry time
	nextDay := parsedTime.AddDate(0, 0, 1)
	expiryTime := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, loc)

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
	now := wsm.TimeManager.Now()
	currentDate := now.Format("2006-01-02")
	cacheKey := makeKey(strings.ToLower(location), currentDate)

	if cachedAstroResponse, found := wsm.Cache[cacheKey]; found {
		if now.Before(cachedAstroResponse.ExpiresAt) {
			log.Println("value served from cache", cachedAstroResponse)
			return cachedAstroResponse, nil
		}
		log.Printf("cached value for %s has expired\n %+v %+v", location, now, cachedAstroResponse.ExpiresAt)
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

	expiryTime, err := wsm.TimeManager.ComputeExpiryTime(forecastResponse.Location.LocalTime, forecastResponse.Location.TimeZoneID)
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
		LocationName:    forecastResponse.Location.Name,
		LocationRegion:  forecastResponse.Location.Region,
		LocationCountry: forecastResponse.Location.Country,
		AstroData:       forecast.AstroData,
		Date:            forecastDate,
		ExpiresAt:       expiryTime,
	}

	wsm.Cache[cacheKey] = astroResponse

	log.Println("value served from fresh API request", astroResponse)
	return astroResponse, nil
}

func makeKey(location string, date string) string {
	return fmt.Sprintf("%s:%s", location, date)
}

func AstroHandler(w http.ResponseWriter, r *http.Request, wm WeatherManager) {
	// Parse location from request's query parameters
	location := r.URL.Query().Get("location")

	if location == "" {
		log.Println("Error: 'location' query parameter is missing in the incoming request.")
		http.Error(w, "Error: 'location' query parameter is missing. Please provide a valid location.", http.StatusBadRequest)
		return
	}

	// Get the astro data
	astroData, err := wm.GetAstro(location)
	if err != nil {
		log.Printf("Error fetching astro data for location '%s': %v", location, err)
		http.Error(w, fmt.Sprintf("Error: failed to get astro data for location '%s'. Please try again later.", location), http.StatusInternalServerError)
		return
	}

	// Convert the astroData to JSON before sending it
	astroDataJson, err := json.Marshal(astroData)
	if err != nil {
		log.Printf("Error converting astro data to JSON for location '%s': %v", location, err)
		http.Error(w, "Error: failed to process astro data. Please try again later.", http.StatusInternalServerError)
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
		TimeManager: &DefaultTimeManager{},
	}

	mux := http.NewServeMux()

	// Wrap the AstroHandler in a function that matches the expected signature
	mux.HandleFunc("/astro", func(w http.ResponseWriter, r *http.Request) {
		AstroHandler(w, r, &wsm)
	})

	// Applying CORS middleware to the mux
	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"}, // frontend default
	}).Handler(mux)

	host := "127.0.0.1"
	port := "8080"
	fmt.Printf("Server listening on %s port %s\n", host, port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), handler))
}
