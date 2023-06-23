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
}

type WeatherApiResponse struct {
	Location Location `json:"location"`
	Current  Current  `json:"current"`
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

type ApiErrorResponse struct {
	Success bool     `json:"success"`
	Error   ApiError `json:"error"`
}

type ApiError struct {
	Code int    `json:"code"`
	Type string `json:"type"`
	Info string `json:"info"`
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

func CelsiusToFahrenheit(celsius int) int {
	return int(float64(celsius)*1.8 + 32)
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
		BaseUrl:     "http://api.weatherstack.com/",
		WEATHER_KEY: WEATHER_KEY,
	}

	data, err := wsm.GetWeather("Vancouver")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(data)
}
