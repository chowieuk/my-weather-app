package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestWeatherStackManager_GetWeather(t *testing.T) {

	type fields struct {
		BaseUrl     string
		WEATHER_KEY string
		Client      http.Client
	}
	type args struct {
		location string
	}

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the WEATHER_KEY from the environment variables
	WEATHER_KEY := os.Getenv("WEATHER_KEY")

	// create a new server serving the appropriate mock JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		location := query.Get("query")
		key := query.Get("access_key")

		var mockResponse []byte
		var err error
		var statusCode int

		switch {
		case key == "":
			mockResponse, err = os.ReadFile("./test_mocks/101_missing_access_key.json")
			statusCode = 401
		case key == "invalid":
			mockResponse, err = os.ReadFile("./test_mocks/101_invalid_access_key.json")
			statusCode = 401
		case location == "":
			mockResponse, err = os.ReadFile("./test_mocks/601_missing_query.json")
			statusCode = 401
		case location == "602_no_results":
			mockResponse, err = os.ReadFile("./test_mocks/602_no_results.json")
			statusCode = 400
		case location == "104_usage_limit_reached":
			mockResponse, err = os.ReadFile("./test_mocks/104_usage_limit_reached.json")
			statusCode = 400
		default:
			mockResponse, err = os.ReadFile("./test_mocks/weatherCurrentResponse.json")
			statusCode = 200
		}

		if err != nil {
			panic(err)
		}

		w.WriteHeader(statusCode)
		_, _ = w.Write(mockResponse)
	}))
	defer server.Close()

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *WeatherApiResponse
		wantErr bool
	}{
		{
			name: "successful case",
			fields: fields{
				BaseUrl:     server.URL,
				WEATHER_KEY: WEATHER_KEY,
			},
			args: args{
				location: "Vancouver",
			},
			want: &WeatherApiResponse{
				Location: Location{
					Name:    "Vancouver",
					Country: "Canada",
					Region:  "British Columbia",
				},
				Current: Current{
					Temperature: 19,
					Fahrenheit:  66,
				},
			},
			wantErr: false,
		},
		{
			name: "101_missing_access_key error",
			fields: fields{
				BaseUrl:     server.URL,
				WEATHER_KEY: "",
			},
			args: args{
				location: "Vancouver",
			},
			wantErr: true,
		},
		{
			name: "101_invalid_access_key error",
			fields: fields{
				BaseUrl:     server.URL,
				WEATHER_KEY: "invalid",
			},
			args: args{
				location: "Vancouver",
			},
			wantErr: true,
		},
		{
			name: "601_missing_query error",
			fields: fields{
				BaseUrl:     server.URL,
				WEATHER_KEY: WEATHER_KEY,
			},
			args: args{
				location: "", // this empty string triggers the 601 error in our mock server
			},
			wantErr: true,
		},
		{
			name: "602_no_results error",
			fields: fields{
				BaseUrl:     server.URL,
				WEATHER_KEY: WEATHER_KEY,
			},
			args: args{
				location: "602_no_results", // this string triggers the 602 error in our mock server
			},
			wantErr: true,
		},
		{
			name: "104_usage_limit_reached error",
			fields: fields{
				BaseUrl:     server.URL,
				WEATHER_KEY: WEATHER_KEY,
			},
			args: args{
				location: "104_usage_limit_reached", // this string triggers the 104 error in our mock server
			},

			wantErr: true,
		},
		{
			name: "server error",
			fields: fields{
				BaseUrl:     "https://fake.unresolved",
				WEATHER_KEY: WEATHER_KEY,
			},
			args: args{
				location: "any location",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wsm := &WeatherStackManager{
				BaseUrl:     tt.fields.BaseUrl,
				WEATHER_KEY: tt.fields.WEATHER_KEY,
				Client:      tt.fields.Client,
			}
			got, err := wsm.GetWeather(tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("WeatherStackManager.GetWeather() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && (got.Location != tt.want.Location || got.Current != tt.want.Current) {
				t.Errorf("WeatherStackManager.GetWeather() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCelsiusToFahrenheit(t *testing.T) {
	result := CelsiusToFahrenheit(0)
	if result != 32 {
		t.Errorf("Expected 32, but got %d", result)
	}

	result = CelsiusToFahrenheit(100)
	if result != 212 {
		t.Errorf("Expected 212, but got %d", result)
	}
}
