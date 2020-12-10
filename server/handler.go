package server

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/gorilla/mux"
)

var (
	ErrUserExists    = errors.New("user already exists")
	ErrUserNotFound  = errors.New("user does not exist")
	ErrBadPassword   = errors.New("password does not match")
	ErrBadAuthHeader = errors.New("bad authentication header format")
	ErrBadAuth       = errors.New("invalid auth token")

	jwtSecret = []byte("UNSECURE_SECRET_TOKEN")
)

// GET /api/airports?search=xxx
type jsonAirportSearchResp struct {
	Data    []jsonAirport `json:"data"`
	Context jsonContext   `json:"context"`
}

func (s *Server) AirportSearch(w http.ResponseWriter, req *http.Request) {
	searchKey := req.FormValue("search")

	respData, err := s.db.GetAirports(searchKey)
	if err != nil {
		writeJsonFailure(w, 500, err)
		return
	}

	encodeRespOrFail(w, respData)
}

// GET /api/flightPaths/{from}/{to}?leave=mm/dd/YYYY
type jsonFlightSearchResp struct {
	Data    []jsonFlight `json:"data"`
	Context jsonContext  `json:"context"`
}

func (s *Server) FlightSearch(w http.ResponseWriter, req *http.Request) {
	reqVars := mux.Vars(req)

	leaveDate, err := time.Parse("01/02/2006", req.FormValue("leave"))
	if err != nil {
		writeJsonFailure(w, 500, err)
		return
	}

	var respData jsonFlightSearchResp
	dayOfWeek := int(leaveDate.Weekday())
	respData, err = s.db.GetFlightPaths(reqVars["from"], reqVars["to"], dayOfWeek)
	if err != nil {
		writeJsonFailure(w, 500, err)
		return
	}

	encodeRespOrFail(w, respData)
}

// GET /api/hotel/{description}/{location}
type jsonHotelSearchResp struct {
	Data    []jsonHotel `json:"data"`
	Context jsonContext `json:"context"`
}

func (s *Server) HotelSearch(w http.ResponseWriter, req *http.Request) {
	reqVars := mux.Vars(req)

	description := reqVars["description"]
	location := reqVars["location"]

	respData, err := s.db.GetHotels(description, location)
	if err != nil {
		writeJsonFailure(w, 500, err)
		return
	}

	encodeRespOrFail(w, respData)
}

// POST /api/user/login
type jsonUserLoginReq struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type jsonUserLoginResp struct {
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
	Context jsonContext `json:"context"`
}

func (s *Server) UserLogin(w http.ResponseWriter, req *http.Request) {
	var respData jsonUserLoginResp

	var reqData jsonUserLoginReq
	if !decodeReqOrFail(w, req, &reqData) {
		return
	}

	password, err := s.db.GetUserPassword(reqData.User)
	if errors.Is(err, gocb.ErrDocumentNotFound) {
		writeJsonFailure(w, 401, ErrUserNotFound)
		return
	} else if err != nil {
		fmt.Println(errors.Unwrap(err))
		writeJsonFailure(w, 500, err)
		return
	}

	if password != reqData.Password {
		writeJsonFailure(w, 401, ErrBadPassword)
		return
	}

	token, err := createJwtToken(reqData.User)
	if err != nil {
		writeJsonFailure(w, 500, err)
		return
	}

	respData.Data.Token = token

	encodeRespOrFail(w, respData)
}

//POST /api/user/signup
type jsonUserSignupReq struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type jsonUserSignupResp struct {
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
	Context jsonContext `json:"context"`
}

func (s *Server) UserSignup(w http.ResponseWriter, req *http.Request) {
	var respData jsonUserSignupResp

	var reqData jsonUserSignupReq
	if !decodeReqOrFail(w, req, &reqData) {
		return
	}

	err := s.db.CreateUser(reqData.User, reqData.Password)
	if errors.Is(err, gocb.ErrDocumentExists) {
		writeJsonFailure(w, 409, ErrUserExists)
		return
	} else if err != nil {
		fmt.Println(errors.Unwrap(err))
		writeJsonFailure(w, 500, err)
		return
	}

	token, err := createJwtToken(reqData.User)
	if err != nil {
		writeJsonFailure(w, 500, err)
		return
	}

	respData.Data.Token = token

	encodeRespOrFail(w, respData)
}

// GET /api/user/{username}/flights
type jsonUserFlightsResp struct {
	Data    []jsonBookedFlight `json:"data"`
	Context jsonContext        `json:"context"`
}

func (s *Server) UserFlights(w http.ResponseWriter, req *http.Request) {
	var respData jsonUserFlightsResp
	var authUser authedUser

	if !decodeAuthUserOrFail(w, req, &authUser) {
		return
	}

	respData, err := s.db.GetUserFlights(authUser.Name)
	if err != nil {
		writeJsonFailure(w, 500, err)
		return
	}

	encodeRespOrFail(w, respData)
}

//POST  /api/user/{username}/flights
type jsonUserBookFlightReq struct {
	Flights []jsonBookedFlight `json:"flights"`
}

type jsonUserBookFlightResp struct {
	Data struct {
		Added []jsonBookedFlight `json:"added"`
	} `json:"data"`
	Context jsonContext `json:"context"`
}

func (s *Server) UserBookFlight(w http.ResponseWriter, req *http.Request) {
	var respData jsonUserBookFlightResp
	var reqData jsonUserBookFlightReq
	var authUser authedUser

	if !decodeAuthUserOrFail(w, req, &authUser) {
		return
	}

	if !decodeReqOrFail(w, req, &reqData) {
		return
	}

	respData, err := s.db.UpdateUserFlights(authUser.Name, reqData.Flights)
	if err != nil {
		writeJsonFailure(w, 500, err)
		return
	}

	encodeRespOrFail(w, respData)
}
