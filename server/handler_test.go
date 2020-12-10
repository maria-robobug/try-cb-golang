package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/google/go-cmp/cmp"
)

func TestAirportSearch(t *testing.T) {
	t.Parallel()

	validData := jsonAirportSearchResp{
		Context: jsonContext{"test"},
		Data:    []jsonAirport{{"San Francisco Intl"}},
	}

	testCases := []struct {
		title      string
		endpoint   string
		repository Repository

		wantStatus int
		wantResp   jsonAirportSearchResp
	}{
		{
			title:    "200 - ok",
			endpoint: "/api/airports?search=SFO",
			repository: &mockRepo{
				GetAirportsFn: func(searchKey string) (jsonAirportSearchResp, error) {
					if searchKey != "SFO" {
						t.Errorf("unexpected search key, got: %s want: %s", searchKey, "SFO")
					}

					return validData, nil
				},
			},

			wantStatus: http.StatusOK,
			wantResp:   validData,
		},
		{
			title:    "500 - error querying data",
			endpoint: "/api/airports?search=boom",
			repository: &mockRepo{
				GetAirportsFn: func(searchKey string) (jsonAirportSearchResp, error) {
					return jsonAirportSearchResp{}, errors.New("boom")
				},
			},

			wantStatus: http.StatusInternalServerError,
			wantResp:   jsonAirportSearchResp{},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tc.endpoint, nil)

			server := New(tc.repository)
			server.ServeHTTP(w, r)

			// Check the status code is what we expect.
			if status := w.Code; status != tc.wantStatus {
				t.Errorf("invalid status code: \ngot: %#v, \nwant: %#v", status, tc.wantStatus)
			}

			// Check the response is what we expect.
			var gotResp jsonAirportSearchResp
			if err := json.Unmarshal(w.Body.Bytes(), &gotResp); err != nil {
				t.Fatal("error unmarshaling json:", err)
			}
			if diff := cmp.Diff(gotResp, tc.wantResp); diff != "" {
				t.Errorf("invalid response body: \ngot: %q, \nwant: %q", gotResp, tc.wantResp)
			}
		})
	}
}

func TestFlightSearch(t *testing.T) {
	t.Parallel()

	leaveDate, err := time.Parse("01/02/2006", "12/15/2020")
	if err != nil {
		t.Fatalf("error parsing date")
	}

	validData := jsonFlightSearchResp{
		Context: jsonContext{"test"},
		Data: []jsonFlight{
			{Name: "FLIGHT1", Flight: "1234HH"},
		},
	}

	testCases := []struct {
		title      string
		endpoint   string
		repository Repository

		wantStatus int
		wantResp   jsonFlightSearchResp
	}{
		{
			title:    "200 - ok",
			endpoint: "/api/flightPaths/airport_a/airport_b?leave=12/15/2020",
			repository: &mockRepo{
				GetFlightPathsFn: func(from, to string, dayOfWeek int) (jsonFlightSearchResp, error) {
					if from != "airport_a" {
						t.Errorf("unexpected from param, got: %s want: %s", from, "airport_a")
					}
					if to != "airport_b" {
						t.Errorf("unexpected to param, got: %s want: %s", to, "airport_b")
					}

					validDay := int(leaveDate.Weekday())
					if dayOfWeek != validDay {
						t.Errorf("unexpected day param, got: %d want: %d", dayOfWeek, validDay)
					}

					return validData, nil
				},
			},

			wantStatus: http.StatusOK,
			wantResp:   validData,
		},
		{
			title:    "500 - invalid leave param",
			endpoint: "/api/flightPaths/boom/boom?leave=",
			repository: &mockRepo{
				GetFlightPathsFn: func(from, to string, dayOfWeek int) (jsonFlightSearchResp, error) {
					return jsonFlightSearchResp{}, nil
				},
			},

			wantStatus: http.StatusInternalServerError,
			wantResp:   jsonFlightSearchResp{},
		},
		{
			title:    "500 - error querying data",
			endpoint: "/api/flightPaths/airport_a/airport_b?leave=12/15/2020",
			repository: &mockRepo{
				GetFlightPathsFn: func(from, to string, dayOfWeek int) (jsonFlightSearchResp, error) {
					return jsonFlightSearchResp{}, errors.New("boom")
				},
			},

			wantStatus: http.StatusInternalServerError,
			wantResp:   jsonFlightSearchResp{},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tc.endpoint, nil)

			server := New(tc.repository)
			server.ServeHTTP(w, r)

			// Check the status code is what we expect.
			if status := w.Code; status != tc.wantStatus {
				t.Errorf("invalid status code: \ngot: %#v, \nwant: %#v", status, tc.wantStatus)
			}

			// Check the response is what we expect.
			var gotResp jsonFlightSearchResp
			if err := json.Unmarshal(w.Body.Bytes(), &gotResp); err != nil {
				t.Fatal("error unmarshaling json:", err)
			}
			if diff := cmp.Diff(gotResp, tc.wantResp); diff != "" {
				t.Errorf("invalid response body: \ngot: %#v, \nwant: %#v", gotResp, tc.wantResp)
			}
		})
	}
}

func TestHotelSearch(t *testing.T) {
	t.Parallel()

	validData := jsonHotelSearchResp{
		Context: jsonContext{"test"},
		Data: []jsonHotel{
			{Country: "UK", Description: "Four Star"},
		},
	}

	testCases := []struct {
		title      string
		endpoint   string
		repository Repository

		wantStatus int
		wantResp   jsonHotelSearchResp
	}{
		{
			title:    "200 - ok with description",
			endpoint: "/api/hotel/Four%20star/",
			repository: &mockRepo{
				GetHotelsFn: func(description, location string) (jsonHotelSearchResp, error) {
					if description != "Four star" {
						t.Errorf("unexpected description param, got: %s want: %s", description, "Four star")
					}

					return validData, nil
				},
			},

			wantStatus: http.StatusOK,
			wantResp:   validData,
		},
		{
			title:    "200 - ok with description and location",
			endpoint: "/api/hotel/Four%20star/London/",
			repository: &mockRepo{
				GetHotelsFn: func(description, location string) (jsonHotelSearchResp, error) {
					if description != "Four star" {
						t.Errorf("unexpected description param, got: %s want: %s", description, "Four star")
					}
					if location != "London" {
						t.Errorf("unexpected location param, got: %s want: %s", location, "London")
					}

					return validData, nil
				},
			},

			wantStatus: http.StatusOK,
			wantResp:   validData,
		},
		{
			title:    "500 - error querying data",
			endpoint: "/api/hotel/boom/",
			repository: &mockRepo{
				GetHotelsFn: func(description, location string) (jsonHotelSearchResp, error) {
					return jsonHotelSearchResp{}, errors.New("boom")
				},
			},

			wantStatus: http.StatusInternalServerError,
			wantResp:   jsonHotelSearchResp{},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tc.endpoint, nil)

			server := New(tc.repository)
			server.ServeHTTP(w, r)

			// Check the status code is what we expect.
			if status := w.Code; status != tc.wantStatus {
				t.Errorf("invalid status code: \ngot: %#v, \nwant: %#v", status, tc.wantStatus)
			}

			// Check the response is what we expect.
			var gotResp jsonHotelSearchResp
			if err := json.Unmarshal(w.Body.Bytes(), &gotResp); err != nil {
				t.Fatal("error unmarshaling json:", err)
			}
			if diff := cmp.Diff(gotResp, tc.wantResp); diff != "" {
				t.Errorf("invalid response body: \ngot: %#v, \nwant: %#v", gotResp, tc.wantResp)
			}
		})
	}
}

func TestUserLogin(t *testing.T) {
	t.Parallel()

	validJwtToken, err := createJwtToken("test_user")
	if err != nil {
		t.Fatal("error creating test jwt token:", err)
	}

	var validData jsonUserLoginResp
	validData.Data.Token = validJwtToken

	testCases := []struct {
		title      string
		endpoint   string
		reqBody    []byte
		repository Repository

		wantStatus int
		wantResp   jsonUserLoginResp
	}{
		{
			title:    "200 - ok valid user",
			endpoint: "/api/user/login",
			reqBody:  []byte(`{"user":"test_user","password":"test_passw"}`),
			repository: &mockRepo{
				GetUserPasswordFn: func(username string) (string, error) {
					if username != "test_user" {
						t.Errorf("unexpected username param, got: %s want: %s", username, "test_user")
					}

					return "test_passw", nil
				},
			},

			wantStatus: http.StatusOK,
			wantResp:   validData,
		},
		{
			title:    "500 - error decoding request",
			endpoint: "/api/user/login",
			reqBody:  []byte(`{"user":}`),
			repository: &mockRepo{
				GetUserPasswordFn: func(username string) (string, error) {
					return "", nil
				},
			},

			wantStatus: http.StatusInternalServerError,
			wantResp:   jsonUserLoginResp{},
		},
		{
			title:    "401 - user not found",
			endpoint: "/api/user/login",
			reqBody:  []byte(`{"user":"test_user","password":"test_passw"}`),
			repository: &mockRepo{
				GetUserPasswordFn: func(username string) (string, error) {
					return "", gocb.ErrDocumentNotFound
				},
			},

			wantStatus: http.StatusUnauthorized,
			wantResp:   jsonUserLoginResp{},
		},
		{
			title:    "500 - error querying data",
			endpoint: "/api/user/login",
			reqBody:  []byte(`{"user":"test_user","password":"test_passw"}`),
			repository: &mockRepo{
				GetUserPasswordFn: func(username string) (string, error) {
					return "", errors.New("boom")
				},
			},

			wantStatus: http.StatusInternalServerError,
			wantResp:   jsonUserLoginResp{},
		},
		{
			title:    "401 - error password mismatch",
			endpoint: "/api/user/login",
			reqBody:  []byte(`{"user":"test_user","password":"test_passw"}`),
			repository: &mockRepo{
				GetUserPasswordFn: func(username string) (string, error) {
					return "boom", nil
				},
			},

			wantStatus: http.StatusUnauthorized,
			wantResp:   jsonUserLoginResp{},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, tc.endpoint, bytes.NewBuffer(tc.reqBody))

			server := New(tc.repository)
			server.ServeHTTP(w, r)

			// Check the status code is what we expect.
			if status := w.Code; status != tc.wantStatus {
				t.Errorf("invalid status code: \ngot: %#v, \nwant: %#v", status, tc.wantStatus)
			}

			// Check the response is what we expect.
			var gotResp jsonUserLoginResp
			if err := json.Unmarshal(w.Body.Bytes(), &gotResp); err != nil {
				t.Fatal("error unmarshaling json:", err)
			}
			if diff := cmp.Diff(gotResp, tc.wantResp); diff != "" {
				t.Errorf("invalid response body: \ngot: %#v, \nwant: %#v", gotResp, tc.wantResp)
			}
		})
	}
}

func TestUserSignUp(t *testing.T) {
	t.Parallel()

	validJwtToken, err := createJwtToken("test_user")
	if err != nil {
		t.Fatal("error creating test jwt token:", err)
	}

	var validData jsonUserSignupResp
	validData.Data.Token = validJwtToken

	testCases := []struct {
		title      string
		endpoint   string
		reqBody    []byte
		repository Repository

		wantStatus int
		wantResp   jsonUserSignupResp
	}{
		{
			title:    "200 - ok valid user",
			endpoint: "/api/user/signup",
			reqBody:  []byte(`{"user":"test_user","password":"test_passw"}`),
			repository: &mockRepo{
				CreateUserFn: func(username, password string) error {
					if username != "test_user" {
						t.Errorf("unexpected username param, got: %s want: %s", username, "test_user")
					}
					if password != "test_passw" {
						t.Errorf("unexpected password param, got: %s want: %s", password, "test_passw")
					}

					return nil
				},
			},

			wantStatus: http.StatusOK,
			wantResp:   validData,
		},
		{
			title:    "500 - error decoding request",
			endpoint: "/api/user/signup",
			reqBody:  []byte(`{"user":}`),
			repository: &mockRepo{
				CreateUserFn: func(username, password string) error {
					return nil
				},
			},

			wantStatus: http.StatusInternalServerError,
			wantResp:   jsonUserSignupResp{},
		},
		{
			title:    "409 - error user already exists",
			endpoint: "/api/user/signup",
			reqBody:  []byte(`{"user":"test_user","password":"test_passw"}`),
			repository: &mockRepo{
				CreateUserFn: func(username, password string) error {
					return gocb.ErrDocumentExists
				},
			},

			wantStatus: http.StatusConflict,
			wantResp:   jsonUserSignupResp{},
		},
		{
			title:    "500 - error querying data",
			endpoint: "/api/user/signup",
			reqBody:  []byte(`{"user":"test_user","password":"test_passw"}`),
			repository: &mockRepo{
				CreateUserFn: func(username, password string) error {
					return errors.New("boom")
				},
			},

			wantStatus: http.StatusInternalServerError,
			wantResp:   jsonUserSignupResp{},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, tc.endpoint, bytes.NewBuffer(tc.reqBody))

			server := New(tc.repository)
			server.ServeHTTP(w, r)

			// Check the status code is what we expect.
			if status := w.Code; status != tc.wantStatus {
				t.Errorf("invalid status code: \ngot: %#v, \nwant: %#v", status, tc.wantStatus)
			}

			// Check the response is what we expect.
			var gotResp jsonUserSignupResp
			if err := json.Unmarshal(w.Body.Bytes(), &gotResp); err != nil {
				t.Fatal("error unmarshaling json:", err)
			}
			if diff := cmp.Diff(gotResp, tc.wantResp); diff != "" {
				t.Errorf("invalid response body: \ngot: %#v, \nwant: %#v", gotResp, tc.wantResp)
			}
		})
	}
}

func TestUserFlights(t *testing.T) {
	t.Parallel()

	validData := jsonUserFlightsResp{
		Context: jsonContext{"test"},
		Data: []jsonBookedFlight{
			{Name: "FLIGHT1", Flight: "1234HH"},
		},
	}

	validJwtToken, err := createJwtToken("test_user")
	if err != nil {
		t.Fatal("error creating test jwt token:", err)
	}

	invalidJwtToken, err := createJwtToken("")
	if err != nil {
		t.Fatal("error creating test jwt token:", err)
	}

	testCases := []struct {
		title      string
		endpoint   string
		token      string
		repository Repository

		wantStatus int
		wantResp   jsonUserFlightsResp
	}{
		{
			title:    "200 - ok valid user",
			endpoint: "/api/user/test_user/flights",
			token:    "Bearer " + validJwtToken,
			repository: &mockRepo{
				GetUserFlightsFn: func(username string) (jsonUserFlightsResp, error) {
					if username != "test_user" {
						t.Errorf("unexpected username param, got: %s want: %s", username, "test_user")
					}

					return validData, nil
				},
			},

			wantStatus: http.StatusOK,
			wantResp:   validData,
		},
		{
			title:    "400 - bad auth header",
			endpoint: "/api/user/test_user/flights",
			token:    "boom",
			repository: &mockRepo{
				GetUserFlightsFn: func(username string) (jsonUserFlightsResp, error) {
					return jsonUserFlightsResp{}, nil
				},
			},

			wantStatus: http.StatusBadRequest,
			wantResp:   jsonUserFlightsResp{},
		},
		{
			title:    "400 - invalid token",
			endpoint: "/api/user/test_user/flights",
			token:    "Bearer boom",
			repository: &mockRepo{
				GetUserFlightsFn: func(username string) (jsonUserFlightsResp, error) {
					return jsonUserFlightsResp{}, nil
				},
			},

			wantStatus: http.StatusBadRequest,
			wantResp:   jsonUserFlightsResp{},
		},
		{
			title:    "400 - bad auth",
			endpoint: "/api/user/boom/flights",
			token:    "Bearer " + invalidJwtToken,
			repository: &mockRepo{
				GetUserFlightsFn: func(username string) (jsonUserFlightsResp, error) {
					return jsonUserFlightsResp{}, nil
				},
			},

			wantStatus: http.StatusBadRequest,
			wantResp:   jsonUserFlightsResp{},
		},
		{
			title:    "500 - error querying data",
			endpoint: "/api/user/test_user/flights",
			token:    "Bearer " + validJwtToken,
			repository: &mockRepo{
				GetUserFlightsFn: func(username string) (jsonUserFlightsResp, error) {
					return jsonUserFlightsResp{}, errors.New("boom")
				},
			},

			wantStatus: http.StatusInternalServerError,
			wantResp:   jsonUserFlightsResp{},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tc.endpoint, nil)
			r.Header.Set("Authorization", tc.token)

			server := New(tc.repository)
			server.ServeHTTP(w, r)

			// Check the status code is what we expect.
			if status := w.Code; status != tc.wantStatus {
				t.Errorf("invalid status code: \ngot: %#v, \nwant: %#v", status, tc.wantStatus)
			}

			// Check the response is what we expect.
			var gotResp jsonUserFlightsResp
			if err := json.Unmarshal(w.Body.Bytes(), &gotResp); err != nil {
				t.Fatal("error unmarshaling json:", err)
			}
			if diff := cmp.Diff(gotResp, tc.wantResp); diff != "" {
				t.Errorf("invalid response body: \ngot: %#v, \nwant: %#v", gotResp, tc.wantResp)
			}
		})
	}
}

func TestUserBookFlight(t *testing.T) {
	t.Parallel()

	validData := jsonUserBookFlightResp{Context: jsonContext{"test"}}
	flights := []jsonBookedFlight{
		{Name: "US Airways", Flight: "US229", SourceAirport: "SFO", DestinationAirport: "LAX", Price: 158.38},
	}
	validData.Data.Added = flights

	validJwtToken, err := createJwtToken("test_user")
	if err != nil {
		t.Fatal("error creating test jwt token:", err)
	}

	invalidJwtToken, err := createJwtToken("")
	if err != nil {
		t.Fatal("error creating test jwt token:", err)
	}

	testCases := []struct {
		title      string
		endpoint   string
		token      string
		reqBody    []byte
		repository Repository

		wantStatus int
		wantResp   jsonUserBookFlightResp
	}{
		{
			title:    "200 - ok valid user",
			endpoint: "/api/user/test_user/flights",
			token:    "Bearer " + validJwtToken,
			reqBody:  []byte(`{"flights":[{"name":"US Airways","flight":"US229","sourceairport":"SFO","destinationairport":"LAX","price":158.38}]}`),
			repository: &mockRepo{
				UpdateUserFlightsFn: func(username string, flights []jsonBookedFlight) (jsonUserBookFlightResp, error) {
					if username != "test_user" {
						t.Errorf("unexpected username param, got: %s want: %s", username, "test_user")
					}
					if diff := cmp.Diff(flights, validData.Data.Added); diff != "" {
						t.Errorf("unexpected flights param: \ngot: %#v, \nwant: %#v", flights, validData.Data.Added)
					}

					return validData, nil
				},
			},

			wantStatus: http.StatusOK,
			wantResp:   validData,
		},
		{
			title:    "400 - bad auth header",
			endpoint: "/api/user/test_user/flights",
			token:    "boom",
			repository: &mockRepo{
				UpdateUserFlightsFn: func(username string, flights []jsonBookedFlight) (jsonUserBookFlightResp, error) {
					return jsonUserBookFlightResp{}, nil
				},
			},

			wantStatus: http.StatusBadRequest,
			wantResp:   jsonUserBookFlightResp{},
		},
		{
			title:    "400 - invalid token",
			endpoint: "/api/user/test_user/flights",
			token:    "Bearer boom",
			repository: &mockRepo{
				UpdateUserFlightsFn: func(username string, flights []jsonBookedFlight) (jsonUserBookFlightResp, error) {
					return jsonUserBookFlightResp{}, nil
				},
			},

			wantStatus: http.StatusBadRequest,
			wantResp:   jsonUserBookFlightResp{},
		},
		{
			title:    "400 - bad auth",
			endpoint: "/api/user/boom/flights",
			token:    "Bearer " + invalidJwtToken,
			repository: &mockRepo{
				UpdateUserFlightsFn: func(username string, flights []jsonBookedFlight) (jsonUserBookFlightResp, error) {
					return jsonUserBookFlightResp{}, nil
				},
			},

			wantStatus: http.StatusBadRequest,
			wantResp:   jsonUserBookFlightResp{},
		},
		{
			title:    "500 - error decoding request",
			endpoint: "/api/user/test_user/flights",
			token:    "Bearer " + validJwtToken,
			reqBody:  []byte(`{"boom":}`),
			repository: &mockRepo{
				UpdateUserFlightsFn: func(username string, flights []jsonBookedFlight) (jsonUserBookFlightResp, error) {
					return jsonUserBookFlightResp{}, nil
				},
			},

			wantStatus: http.StatusInternalServerError,
			wantResp:   jsonUserBookFlightResp{},
		},
		{
			title:    "500 - error querying data",
			endpoint: "/api/user/test_user/flights",
			reqBody:  []byte(`{"flights":[{"name":"US Airways","flight":"US229","sourceairport":"SFO","destinationairport":"LAX","price":158.38}]}`),
			token:    "Bearer " + validJwtToken,
			repository: &mockRepo{
				UpdateUserFlightsFn: func(username string, flights []jsonBookedFlight) (jsonUserBookFlightResp, error) {
					return jsonUserBookFlightResp{}, errors.New("boom")
				},
			},

			wantStatus: http.StatusInternalServerError,
			wantResp:   jsonUserBookFlightResp{},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, tc.endpoint, bytes.NewBuffer(tc.reqBody))
			r.Header.Set("Authorization", tc.token)

			server := New(tc.repository)
			server.ServeHTTP(w, r)

			// Check the status code is what we expect.
			if status := w.Code; status != tc.wantStatus {
				t.Errorf("invalid status code: \ngot: %#v, \nwant: %#v", status, tc.wantStatus)
			}

			// Check the response is what we expect.
			var gotResp jsonUserBookFlightResp
			if err := json.Unmarshal(w.Body.Bytes(), &gotResp); err != nil {
				t.Fatal("error unmarshaling json:", err)
			}
			if diff := cmp.Diff(gotResp, tc.wantResp); diff != "" {
				t.Errorf("invalid response body: \ngot: %#v, \nwant: %#v", gotResp, tc.wantResp)
			}
		})
	}
}
