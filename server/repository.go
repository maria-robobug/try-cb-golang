package server

import (
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/couchbase/gocb/v2/search"
	"github.com/google/uuid"
)

type Repository interface {
	GetAirports(string) (jsonAirportSearchResp, error)
	GetFlightPaths(string, string, int) (jsonFlightSearchResp, error)
	GetHotels(string, string) (jsonHotelSearchResp, error)

	CreateUser(string, string) error
	GetUserPassword(string) (string, error)
	GetUserFlights(string) (jsonUserFlightsResp, error)
	UpdateUserFlights(string, []jsonBookedFlight) (jsonUserBookFlightResp, error)
}

type CBRepository struct {
	cluster       *gocb.Cluster
	defaultBucket *gocb.Bucket
	userBucket    *gocb.Bucket
}

func NewCBRepository() (*CBRepository, error) {
	var (
		cbConnStr    = "couchbase://localhost"
		cbDataBucket = "travel-sample"
		cbUserBucket = "travel-users"
		cbUsername   = "Administrator"
		cbPassword   = "password"
	)

	clusterOpts := gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: cbUsername,
			Password: cbPassword,
		},
	}

	cluster, err := gocb.Connect(cbConnStr, clusterOpts)
	if err != nil {
		return nil, err
	}

	return &CBRepository{
		cluster:       cluster,
		defaultBucket: cluster.Bucket(cbDataBucket),
		userBucket:    cluster.Bucket(cbUserBucket),
	}, nil
}

// GetAirports returns all airports matching the search key provided
func (cr *CBRepository) GetAirports(searchKey string) (jsonAirportSearchResp, error) {
	var queryStr string
	queryParams := make([]interface{}, 1)

	sameCase := (strings.ToUpper(searchKey) == searchKey || strings.ToLower(searchKey) == searchKey)
	if sameCase && len(searchKey) == 3 {
		// FAA code
		queryParams[0] = strings.ToUpper(searchKey)
		queryStr = "SELECT airportname FROM `travel-sample` WHERE faa=$1"
	} else if sameCase && len(searchKey) == 4 {
		// ICAO code
		queryParams[0] = strings.ToUpper(searchKey)
		queryStr = "SELECT airportname FROM `travel-sample` WHERE icao=$1"
	} else {
		// Airport name
		queryParams[0] = strings.ToLower(searchKey)
		queryStr = "SELECT airportname FROM `travel-sample` WHERE POSITION(LOWER(airportname), $1) = 0"
	}

	var respData jsonAirportSearchResp
	respData.Context.Add(queryStr)
	rows, err := cr.cluster.Query(queryStr, &gocb.QueryOptions{PositionalParameters: queryParams})
	if err != nil {
		return jsonAirportSearchResp{}, err
	}

	respData.Data = []jsonAirport{}
	var airport jsonAirport
	for rows.Next() {
		if err = rows.Row(&airport); err != nil {
			return jsonAirportSearchResp{}, err
		}

		respData.Data = append(respData.Data, airport)
		airport = jsonAirport{}
	}
	if err = rows.Close(); err != nil {
		return jsonAirportSearchResp{}, err
	}

	return respData, nil
}

func (cr *CBRepository) GetFlightPaths(from, to string, dayOfWeek int) (jsonFlightSearchResp, error) {
	var respData jsonFlightSearchResp
	queryParams := make(map[string]interface{}, 1)

	// Find aiport faa code for source and destination airports
	queryParams["fromAirport"] = from
	queryParams["toAirport"] = to
	queryStr :=
		"SELECT faa AS fromFaa FROM `travel-sample`" +
			" WHERE airportname=$fromAirport" +
			" UNION" +
			" SELECT faa AS toFaa FROM `travel-sample`" +
			" WHERE airportname=$toAirport;"

	respData.Context.Add(queryStr)
	var airportInfo jsonAirportInfo
	rows, err := cr.cluster.Query(queryStr, &gocb.QueryOptions{NamedParameters: queryParams})
	if err != nil {
		return jsonFlightSearchResp{}, err
	}

	for rows.Next() {
		if err = rows.Row(&airportInfo); err != nil {
			return jsonFlightSearchResp{}, err
		}
	}
	if err = rows.Close(); err != nil {
		return jsonFlightSearchResp{}, err
	}

	// Search for flights
	queryParams["fromFaa"] = airportInfo.FromFaa
	queryParams["toFaa"] = airportInfo.ToFaa
	queryParams["dayOfWeek"] = dayOfWeek
	queryStr =
		"SELECT a.name, s.flight, s.utc, r.sourceairport, r.destinationairport, r.equipment" +
			" FROM `travel-sample` AS r" +
			" UNNEST r.schedule AS s" +
			" JOIN `travel-sample` AS a ON KEYS r.airlineid" +
			" WHERE r.sourceairport=$fromFaa" +
			" AND r.destinationairport=$toFaa" +
			" AND s.day=$dayOfWeek" +
			" ORDER BY a.name ASC;"

	respData.Context.Add(queryStr)
	rows, err = cr.cluster.Query(queryStr, &gocb.QueryOptions{NamedParameters: queryParams})
	if err != nil {
		return jsonFlightSearchResp{}, err
	}

	respData.Data = []jsonFlight{}
	var flight jsonFlight
	for rows.Next() {
		if err = rows.Row(&flight); err != nil {
			return jsonFlightSearchResp{}, err
		}

		flight.FlightTime = int(math.Ceil(rand.Float64() * 8000))
		flight.Price = math.Ceil(float64(flight.FlightTime)/8*100) / 100
		respData.Data = append(respData.Data, flight)
		flight = jsonFlight{}
	}
	if err = rows.Close(); err != nil {
		return jsonFlightSearchResp{}, err
	}

	return respData, nil
}

func (cr *CBRepository) GetHotels(description, location string) (jsonHotelSearchResp, error) {
	var respData jsonHotelSearchResp
	var defaultCollection = cr.defaultBucket.DefaultCollection()

	qp := search.NewConjunctionQuery(search.NewTermQuery("hotel").Field("type"))

	if location != "" && location != "*" {
		qp.And(search.NewDisjunctionQuery(
			search.NewMatchPhraseQuery(location).Field("country"),
			search.NewMatchPhraseQuery(location).Field("city"),
			search.NewMatchPhraseQuery(location).Field("state"),
			search.NewMatchPhraseQuery(location).Field("address"),
		))
	}

	if description != "" && description != "*" {
		qp.And(search.NewDisjunctionQuery(
			search.NewMatchPhraseQuery(description).Field("description"),
			search.NewMatchPhraseQuery(description).Field("name"),
		))
	}

	results, err := cr.cluster.SearchQuery("hotels", qp, &gocb.SearchOptions{Limit: 100})
	if err != nil {
		return jsonHotelSearchResp{}, err
	}

	respData.Data = []jsonHotel{}
	for results.Next() {
		res, _ := defaultCollection.LookupIn(results.Row().ID, []gocb.LookupInSpec{
			gocb.GetSpec("country", nil),
			gocb.GetSpec("city", nil),
			gocb.GetSpec("state", nil),
			gocb.GetSpec("address", nil),
			gocb.GetSpec("name", nil),
			gocb.GetSpec("description", nil),
		}, nil)
		// We ignore errors here since some hotels are missing various
		//  pieces of data, but every key exists since it came from FTS.

		var hotel jsonHotel
		res.ContentAt(0, &hotel.Country)
		res.ContentAt(1, &hotel.City)
		res.ContentAt(2, &hotel.State)
		res.ContentAt(3, &hotel.Address)
		res.ContentAt(4, &hotel.Name)
		res.ContentAt(5, &hotel.Description)

		respData.Data = append(respData.Data, hotel)
	}

	return respData, nil
}

func (cr *CBRepository) GetUserPassword(username string) (string, error) {
	userDataScope := cr.userBucket.Scope("userData")
	userCollection := userDataScope.Collection("users")

	res, err := userCollection.LookupIn(username, []gocb.LookupInSpec{
		gocb.GetSpec("password", nil),
	}, nil)
	if err != nil {
		return "", err
	}

	var password string
	if err = res.ContentAt(0, &password); err != nil {
		return "", err
	}

	return password, nil
}

func (cr *CBRepository) GetUserFlights(username string) (jsonUserFlightsResp, error) {
	var respData jsonUserFlightsResp

	userDataScope := cr.userBucket.Scope("userData")
	userCollection := userDataScope.Collection("users")
	flightCollection := userDataScope.Collection("flights")

	var flightIDs []string
	res, err := userCollection.LookupIn(username, []gocb.LookupInSpec{
		gocb.GetSpec("flights", nil),
	}, nil)
	if err != nil {
		return jsonUserFlightsResp{}, err
	}

	res.ContentAt(0, &flightIDs)

	var flight jsonBookedFlight
	respData.Data = []jsonBookedFlight{}
	for _, flightID := range flightIDs {
		res, err := flightCollection.Get(flightID, nil)
		if err != nil {
			return jsonUserFlightsResp{}, err
		}

		res.Content(&flight)
		respData.Data = append(respData.Data, flight)
	}

	return respData, nil
}

func (cr *CBRepository) CreateUser(username, password string) error {
	userDataScope := cr.userBucket.Scope("userData")
	userCollection := userDataScope.Collection("users")

	user := jsonUser{
		Name:     username,
		Password: password,
		Flights:  nil,
	}
	if _, err := userCollection.Insert(username, user, nil); err != nil {
		return err
	}

	return nil
}

func (cr *CBRepository) UpdateUserFlights(username string, bookedFlights []jsonBookedFlight) (jsonUserBookFlightResp, error) {
	var respData jsonUserBookFlightResp

	userDataScope := cr.userBucket.Scope("userData")
	userCollection := userDataScope.Collection("users")
	flightCollection := userDataScope.Collection("flights")

	var user jsonUser
	res, err := userCollection.Get(username, nil)
	if err != nil {
		return jsonUserBookFlightResp{}, err
	}

	cas := res.Cas()
	res.Content(&user)

	for _, flight := range bookedFlights {
		flight.BookedOn = time.Now().Format("01/02/2006")
		respData.Data.Added = append(respData.Data.Added, flight)

		flightID, err := uuid.NewRandom()
		if err != nil {
			return jsonUserBookFlightResp{}, err
		}

		user.Flights = append(user.Flights, flightID.String())
		_, err = flightCollection.Upsert(flightID.String(), flight, nil)
		if err != nil {
			return jsonUserBookFlightResp{}, err
		}
	}

	opts := gocb.ReplaceOptions{Cas: cas}
	_, err = userCollection.Replace(username, user, &opts)
	if err != nil {
		// We intentionally do not handle CAS mismatch, as if the users
		//  account was already modified, they probably want to know.
		return jsonUserBookFlightResp{}, err
	}

	return respData, nil
}
