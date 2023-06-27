package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	// "time"

	"github.com/joho/godotenv"
)

type MockWeatherManager struct {
	TimeManager TimeManager
}

func (mwm *MockWeatherManager) GetAstro(location string) (AstroResponse, error) {
	return AstroResponse{
		Date:      mwm.TimeManager.Now().Format("2006-01-02"),
		ExpiresAt: mwm.TimeManager.Now().Add(time.Hour * 2),
		AstroData: AstroData{
			Sunrise:          "05:07 AM",
			Sunset:           "09:22 PM",
			Moonrise:         "06:58 AM",
			Moonset:          "11:52 PM",
			MoonPhase:        "Waxing Crescent",
			MoonIllumination: 3},
	}, nil
}

type MockTimeManager struct {
	NowFunc               func() time.Time
	ComputeExpiryTimeFunc func(localTime string) (time.Time, error)
}

func (tm *MockTimeManager) Now() time.Time {
	if tm.NowFunc != nil {
		return tm.NowFunc()
	}
	return time.Now()
}

func (tm *MockTimeManager) ComputeExpiryTime(localTime string) (time.Time, error) {
	if tm.ComputeExpiryTimeFunc != nil {
		return tm.ComputeExpiryTimeFunc(localTime)
	}
	// default logic if ComputeExpiryTimeFunc is not set
	lt, err := time.Parse("2006-01-02 15:04", localTime)
	if err != nil {
		return time.Time{}, err
	}

	nextDay := lt.AddDate(0, 0, 1)
	expiryTime := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, lt.Location())

	return expiryTime, nil
}

func TestWeatherStackManager_GetWeather(t *testing.T) {

	type fields struct {
		BaseUrl     string
		WEATHER_KEY string
		Client      http.Client
		TimeManager TimeManager
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
					Name:      "Vancouver",
					Country:   "Canada",
					Region:    "British Columbia",
					LocalTime: "2023-06-21 16:30",
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
				TimeManager: &DefaultTimeManager{},
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

func TestWeatherStackManager_GetAstro(t *testing.T) {
	type fields struct {
		BaseUrl     string
		WEATHER_KEY string
		Client      http.Client
		Cache       map[string]AstroData
		TimeManager TimeManager
	}
	type args struct {
		location string
		date     string
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
			mockResponse, err = os.ReadFile("./test_mocks/weatherForecastResponse.json")
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
		want    AstroResponse
		wantErr bool
	}{
		{
			name: "successful case",
			fields: fields{
				BaseUrl:     server.URL,
				WEATHER_KEY: WEATHER_KEY,
				TimeManager: &MockTimeManager{
					NowFunc: func() time.Time {
						// return a fixed time for this test case
						return time.Date(2023, 6, 22, 12, 35, 30, 0, time.UTC)
					},
					ComputeExpiryTimeFunc: func(localTime string) (time.Time, error) {
						// Compute the expiry time based on localTime
						lt, err := time.Parse("2006-01-02 15:04", localTime)
						if err != nil {
							return time.Time{}, err
						}

						nextDay := lt.AddDate(0, 0, 1)
						expiryTime := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, lt.Location())

						return expiryTime, nil
					},
				},
			},
			args: args{
				location: "Vancouver",
			},
			want: AstroResponse{
				Date:      "2023-06-20",
				ExpiresAt: time.Date(2023, 6, 22, 0, 0, 0, 0, time.UTC), // based on the MockTimeManager implementation
				AstroData: AstroData{
					Sunrise:          "05:07 AM",
					Sunset:           "09:22 PM",
					Moonrise:         "06:58 AM",
					Moonset:          "11:52 PM",
					MoonPhase:        "Waxing Crescent",
					MoonIllumination: 3},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wsm := &WeatherStackManager{
				BaseUrl:     tt.fields.BaseUrl,
				WEATHER_KEY: tt.fields.WEATHER_KEY,
				Client:      tt.fields.Client,
				Cache:       make(map[string]AstroResponse),
				TimeManager: tt.fields.TimeManager,
			}
			got, err := wsm.GetAstro(tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("WeatherStackManager.GetForecast() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WeatherStackManager.GetForecast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWeatherStackManager_CachingBehavior(t *testing.T) {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the WEATHER_KEY from the environment variables
	WEATHER_KEY := os.Getenv("WEATHER_KEY")

	requestCount := 0 // Create a counter variable

	// create a new server serving the appropriate mock JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
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
			mockResponse, err = os.ReadFile("./test_mocks/weatherForecastResponse.json")
			statusCode = 200
		}

		if err != nil {
			panic(err)
		}

		w.WriteHeader(statusCode)
		_, _ = w.Write(mockResponse)
	}))
	defer server.Close()

	fixedTime := time.Date(2023, 6, 27, 22, 00, 00, 0, time.UTC)
	timeManager := &MockTimeManager{
		NowFunc: func() time.Time {
			return fixedTime
		},
	}

	wsm := WeatherStackManager{
		BaseUrl:     server.URL,
		WEATHER_KEY: WEATHER_KEY,
		Cache:       make(map[string]AstroResponse),
		TimeManager: timeManager,
	}

	// Call GetAstro to populate the cache
	firstAstro, err := wsm.GetAstro("Vancouver")
	if err != nil {
		t.Fatalf("Unexpected error during GetAstro: %v", err)
	}

	// Call GetAstro again and check that it returns the same data
	secondAstro, err := wsm.GetAstro("Vancouver")
	if err != nil {
		t.Fatalf("Unexpected error during second GetAstro: %v", err)
	}
	if !reflect.DeepEqual(firstAstro, secondAstro) {
		t.Errorf("Second GetAstro returned different data: first %v, second %v", firstAstro, secondAstro)
	}

	// Check that only one request was made to the server
	if requestCount != 1 {
		t.Errorf("Expected 1 request to the server, but got %d", requestCount)
	}

	// Verify the cache directly
	cacheKey := fmt.Sprintf("Vancouver:%s", fixedTime.Format("2006-01-02"))
	cachedData, found := wsm.Cache[cacheKey]
	if !found {
		t.Errorf("No data found in cache for key: %s", cacheKey)
	}

	// Compare cachedAstro with the expected data
	expectedAstro := AstroResponse{
		Date:      "2023-06-20",
		ExpiresAt: time.Date(2023, 6, 22, 0, 0, 0, 0, time.UTC), // based on the MockTimeManager implementation
		AstroData: AstroData{
			Sunrise:          "05:07 AM",
			Sunset:           "09:22 PM",
			Moonrise:         "06:58 AM",
			Moonset:          "11:52 PM",
			MoonPhase:        "Waxing Crescent",
			MoonIllumination: 3,
		},
	}

	if !reflect.DeepEqual(cachedData, expectedAstro) {
		t.Errorf("Cached data is not as expected. Got %v, want %v", cachedData, expectedAstro)
	}

	wsm = WeatherStackManager{
		BaseUrl:     server.URL,
		WEATHER_KEY: WEATHER_KEY,
		Cache:       make(map[string]AstroResponse),
		TimeManager: &MockTimeManager{
			NowFunc: func() time.Time {
				return time.Now()
			},
			ComputeExpiryTimeFunc: func(localTime string) (time.Time, error) {
				// Compute the expiry time based on localTime
				lt, err := time.Parse("2006-01-02 15:04", localTime)
				if err != nil {
					return time.Time{}, err
				}

				nextDay := lt.AddDate(0, 0, 1)
				expiryTime := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, time.UTC)

				// Return a time in the past
				return expiryTime.Add(-time.Hour), nil
			},
		},
	}

	// Call GetAstro to populate the cache
	thirdAstro, err := wsm.GetAstro("Vancouver")
	if err != nil {
		t.Fatalf("Unexpected error during third GetAstro: %v", err)
	}

	// Call GetAstro again and check that it triggers another request (since the cache has expired)
	fourthAstro, err := wsm.GetAstro("Vancouver")
	if err != nil {
		t.Fatalf("Unexpected error during fourth GetAstro: %v", err)
	}

	// Because the cache is expired, the server should receive a new request, hence increasing the requestCount
	if requestCount != 2 { // two 2 previous tests each making one request
		t.Errorf("Expected 2 requests to the server, but got %d", requestCount)
	}

	// Check that the two responses are the same, despite the cache having been invalidated
	if !reflect.DeepEqual(thirdAstro, fourthAstro) {
		t.Errorf("Second GetAstro returned different data: first %v, second %v", thirdAstro, fourthAstro)
	}
}

func TestAstroHandler(t *testing.T) {

	mwm := &MockWeatherManager{
		TimeManager: &MockTimeManager{
			NowFunc: func() time.Time {
				// return a fixed time for this test case
				return time.Date(2023, 6, 27, 22, 00, 00, 0, time.UTC)
			},
			ComputeExpiryTimeFunc: func(localTime string) (time.Time, error) {
				// Compute the expiry time based on localTime
				lt, err := time.Parse("2006-01-02 15:04", localTime)
				if err != nil {
					return time.Time{}, err
				}

				nextDay := lt.AddDate(0, 0, 1)
				expiryTime := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, time.UTC)

				return expiryTime, nil
			},
		},
	}

	req, err := http.NewRequest("GET", "/astro?location=Vancouver&date=2023-06-20", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	AstroHandler(rr, req, mwm)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"date":"2023-06-27","astro":{"sunrise":"05:07 AM","sunset":"09:22 PM","moonrise":"06:58 AM","moonset":"11:52 PM","moon_phase":"Waxing Crescent","moon_illumination":3},"expires_at":"2023-06-28T00:00:00Z"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
