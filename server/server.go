package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type Server struct {
	router *mux.Router
	db     Repository
}

func New(db Repository) *Server {
	s := &Server{}
	s.setupRoutes()
	s.db = db

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

func (s *Server) setupRoutes() {
	// Create a router for our server
	s.router = mux.NewRouter()

	// Set up our REST endpoints
	s.router.Path("/api/airports").Methods("GET").HandlerFunc(s.AirportSearch)
	s.router.Path("/api/flightPaths/{from}/{to}").Methods("GET").HandlerFunc(s.FlightSearch)
	s.router.Path("/api/user/login").Methods("POST").HandlerFunc(s.UserLogin)
	s.router.Path("/api/user/signup").Methods("POST").HandlerFunc(s.UserSignup)
	s.router.Path("/api/user/{username}/flights").Methods("GET").HandlerFunc(s.UserFlights)
	s.router.Path("/api/user/{username}/flights").Methods("POST").HandlerFunc(s.UserBookFlight)
	s.router.Path("/api/hotel/{description}/").Methods("GET").HandlerFunc(s.HotelSearch)
	s.router.Path("/api/hotel/{description}/{location}/").Methods("GET").HandlerFunc(s.HotelSearch)

	// Serve our public files out of root
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))
}

func writeJsonFailure(w http.ResponseWriter, code int, err error) {
	failObj := struct {
		Failure string `json:"failure"`
	}{
		Failure: err.Error(),
	}

	failBytes, err := json.Marshal(failObj)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(code)
	w.Write(failBytes)
}

func decodeReqOrFail(w http.ResponseWriter, req *http.Request, data interface{}) bool {
	err := json.NewDecoder(req.Body).Decode(data)
	if err != nil {
		writeJsonFailure(w, 500, err)
		return false
	}
	return true
}

func encodeRespOrFail(w http.ResponseWriter, data interface{}) {
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		writeJsonFailure(w, 500, err)
	}
}

func createJwtToken(user string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
	}).SignedString(jwtSecret)
}

func decodeAuthUserOrFail(w http.ResponseWriter, req *http.Request, user *authedUser) bool {
	authHeader := req.Header.Get("Authorization")
	authHeaderParts := strings.SplitN(authHeader, " ", 2)
	if authHeaderParts[0] != "Bearer" {
		authHeader = req.Header.Get("Authentication")
		authHeaderParts = strings.SplitN(authHeader, " ", 2)
		if authHeaderParts[0] != "Bearer" {
			writeJsonFailure(w, 400, ErrBadAuthHeader)
			return false
		}
	}

	authToken := authHeaderParts[1]
	token, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return jwtSecret, nil
	})
	if err != nil {
		writeJsonFailure(w, 400, ErrBadAuthHeader)
		return false
	}

	authUser := token.Claims.(jwt.MapClaims)["user"].(string)
	if authUser == "" {
		writeJsonFailure(w, 400, ErrBadAuth)
		return false
	}

	user.Name = authUser

	return true
}
