package server

type jsonAirport struct {
	AirportName string `json:"airportname"`
}

type jsonAirportInfo struct {
	FromFaa string `json:"fromFaa"`
	ToFaa   string `json:"toFaa"`
}

type jsonFlight struct {
	Name               string  `json:"name"`
	Flight             string  `json:"flight"`
	Equipment          string  `json:"equipment"`
	Utc                string  `json:"utc"`
	SourceAirport      string  `json:"sourceairport"`
	DestinationAirport string  `json:"destinationairport"`
	Price              float64 `json:"price"`
	FlightTime         int     `json:"flighttime"`
}

type jsonHotel struct {
	Country     string `json:"country"`
	City        string `json:"city"`
	State       string `json:"state"`
	Address     string `json:"address"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type jsonBookedFlight struct {
	Name               string  `json:"name"`
	Flight             string  `json:"flight"`
	Price              float64 `json:"price"`
	Date               string  `json:"date"`
	SourceAirport      string  `json:"sourceairport"`
	DestinationAirport string  `json:"destinationairport"`
	BookedOn           string  `json:"bookedon"`
}

type jsonUser struct {
	Name     string   `json:"name"`
	Password string   `json:"password"`
	Flights  []string `json:"flights"`
}

type authedUser struct {
	Name string
}

type jsonContext []string

func (c *jsonContext) Add(msg string) {
	*c = append(*c, msg)
}
