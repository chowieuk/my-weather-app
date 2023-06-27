package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type WeatherStackManager struct {
	BaseUrl     string
	WEATHER_KEY string
	Client      http.Client
	Cache       map[string]Astro
}

type WeatherApiResponse struct {
	Location Location            `json:"location"`
	Current  Current             `json:"current"`
	Forecast map[string]Forecast `json:"forecast"`
}

type Location struct {
	Name    string `json:"name"`
	Country string `json:"country"`
	Region  string `json:"region"`
}

type Current struct {
	Temperature int `json:"temperature"`
	Fahrenheit  int `json:"fahrenheit"`
}

type Forecast struct {
	Date      string  `json:"date"`
	DateEpoch int64   `json:"date_epoch"`
	Astro     Astro   `json:"astro"`
	MinTemp   int     `json:"mintemp"`
	MaxTemp   int     `json:"maxtemp"`
	AvgTemp   int     `json:"avgtemp"`
	TotalSnow int     `json:"totalsnow"`
	SunHour   float64 `json:"sunhour"`
	UVIndex   int     `json:"uv_index"`
}

type Astro struct {
	Sunrise          string `json:"sunrise"`
	Sunset           string `json:"sunset"`
	Moonrise         string `json:"moonrise"`
	Moonset          string `json:"moonset"`
	MoonPhase        string `json:"moon_phase"`
	MoonIllumination int    `json:"moon_illumination"`
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

func makeKey(location string, date string) string {
	return fmt.Sprintf("%s:%s", location, date)
}

func (wsm *WeatherStackManager) GetWeather(location string) (*WeatherApiResponse, error) {
	url := fmt.Sprintf("%s/current?access_key=%s&query=%s", wsm.BaseUrl, wsm.WEATHER_KEY, location)
	response, err := wsm.Client.Get(url)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		var apiErrorResponse ApiErrorResponse
		err = json.Unmarshal(body, &apiErrorResponse)
		if err != nil {
			return nil, err
		}
		log.Printf("%+v\n", &apiErrorResponse)
		return nil, fmt.Errorf("API error occurred: code=%d, type=%s, info=%s", apiErrorResponse.Error.Code, apiErrorResponse.Error.Type, apiErrorResponse.Error.Info)
	}

	var weatherResponse WeatherApiResponse
	err = json.Unmarshal(body, &weatherResponse)
	if err != nil {
		return nil, err
	}

	weatherResponse.Current.Fahrenheit = CelsiusToFahrenheit(weatherResponse.Current.Temperature)

	log.Printf("%+v\n", &weatherResponse)
	return &weatherResponse, nil
}

func (wsm *WeatherStackManager) GetAstro(location string, date string) (Astro, error) {

	// Check the cache first
	key := makeKey(location, date)
	cachedData, found := wsm.Cache[key]
	if found {
		return cachedData, nil
	}

	url := fmt.Sprintf("%s/forecast?access_key=%s&query=%s", wsm.BaseUrl, wsm.WEATHER_KEY, location)
	response, err := wsm.Client.Get(url)
	if err != nil {
		log.Println(err)
		return Astro{}, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return Astro{}, err
	}

	if response.StatusCode != http.StatusOK {
		var apiErrorResponse ApiErrorResponse
		err = json.Unmarshal(body, &apiErrorResponse)
		if err != nil {
			log.Println(err)
			return Astro{}, err
		}
		log.Printf("%+v\n", &apiErrorResponse)
		return Astro{}, fmt.Errorf("API error occurred: code=%d, type=%s, info=%s", apiErrorResponse.Error.Code, apiErrorResponse.Error.Type, apiErrorResponse.Error.Info)
	}

	var forecastResponse WeatherApiResponse
	err = json.Unmarshal(body, &forecastResponse)
	if err != nil {
		log.Println(err)
		return Astro{}, err
	}

	forecastData := forecastResponse.Forecast[date]
	astroData := forecastData.Astro
	wsm.Cache[makeKey(location, date)] = astroData

	return astroData, nil
}

func CelsiusToFahrenheit(celsius int) int {
	return int(float64(celsius)*1.8 + 32)
}

type WeatherManager interface {
	GetAstro(location string, date string) (Astro, error)
}

func AstroHandler(w http.ResponseWriter, r *http.Request, wm WeatherManager) {
	// Parse location and date from request's query parameters
	location := r.URL.Query().Get("location")
	date := r.URL.Query().Get("date")

	if location == "" || date == "" {
		http.Error(w, "Missing 'location' or 'date' query parameter", http.StatusBadRequest)
		return
	}

	// Get the astro data
	astroData, err := wm.GetAstro(location, date)
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
		Cache:       make(map[string]Astro),
	}

	// Wrap the AstroHandler in a function that matches the expected signature
	http.HandleFunc("/astro", func(w http.ResponseWriter, r *http.Request) {
		AstroHandler(w, r, &wsm)
	})
	fmt.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
