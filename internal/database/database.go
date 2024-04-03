package database

type GeoNames struct {
	En string `json:"en"`
	Ru string `json:"ru"`
}

type GeoContinent struct {
	Code  string   `json:"code"`
	Names GeoNames `json:"names"`
}

type GeoCountry struct {
	ISOCode string   `json:"iso_code"`
	Names   GeoNames `json:"names"`
}

type GeoSubdivision struct {
	ISOCode string   `json:"iso_code"`
	Names   GeoNames `json:"names"`
}

type GeoCity struct {
	Names GeoNames `json:"names"`
}

type GeoLocation struct {
	AccuracyRadius int32   `json:"accuracy_radius,omitempty"` // Радиус точности (MaxMind)
	Latitude       float32 `json:"latitude"`                  // Широта
	Longitude      float32 `json:"longitude"`                 // Долгота
	TimeZone       string  `json:"time_zone"`                 // Часовой пояс
}

type GeoRecord struct {
	Continent    GeoContinent     `json:"continent"`    // Континент
	Country      GeoCountry       `json:"country"`      // Страна
	Subdivisions []GeoSubdivision `json:"subdivisions"` // Территориальные подразделения
	City         GeoCity          `json:"city"`         // Город
	Location     GeoLocation      `json:"location"`     // Местоположение
}

type DB interface {
	Lookup(ip string) (*GeoRecord, error)
}
