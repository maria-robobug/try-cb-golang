package server

type mockRepository interface {
	GetAirports(string) (jsonAirportSearchResp, error)
	GetFlightPaths(string, string, int) (jsonFlightSearchResp, error)
	GetHotels(string, string) (jsonHotelSearchResp, error)

	CreateUser(string, string) error
	GetUserPassword(string) (string, error)
	GetUserFlights(string) (jsonUserFlightsResp, error)
	UpdateUserFlights(string, []jsonBookedFlight) (jsonUserBookFlightResp, error)
}

type mockRepo struct {
	GetAirportsFn    func(string) (jsonAirportSearchResp, error)
	GetFlightPathsFn func(string, string, int) (jsonFlightSearchResp, error)
	GetHotelsFn      func(string, string) (jsonHotelSearchResp, error)

	CreateUserFn        func(string, string) error
	GetUserPasswordFn   func(string) (string, error)
	GetUserFlightsFn    func(string) (jsonUserFlightsResp, error)
	UpdateUserFlightsFn func(string, []jsonBookedFlight) (jsonUserBookFlightResp, error)
}

func (mr *mockRepo) GetAirports(searchKey string) (jsonAirportSearchResp, error) {
	return mr.GetAirportsFn(searchKey)
}

func (mr *mockRepo) GetFlightPaths(from, to string, dayOfWeek int) (jsonFlightSearchResp, error) {
	return mr.GetFlightPathsFn(from, to, dayOfWeek)
}

func (mr *mockRepo) GetHotels(description, location string) (jsonHotelSearchResp, error) {
	return mr.GetHotelsFn(description, location)
}

func (mr *mockRepo) GetUserPassword(username string) (string, error) {
	return mr.GetUserPasswordFn(username)
}

func (mr *mockRepo) GetUserFlights(username string) (jsonUserFlightsResp, error) {
	return mr.GetUserFlightsFn(username)
}

func (mr *mockRepo) CreateUser(username, password string) error {
	return mr.CreateUserFn(username, password)
}

func (mr *mockRepo) UpdateUserFlights(username string, bookedFlights []jsonBookedFlight) (jsonUserBookFlightResp, error) {
	return mr.UpdateUserFlightsFn(username, bookedFlights)
}
