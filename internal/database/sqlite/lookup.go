package sqlite

import (
	"database/sql"
	"encoding/binary"
	"errors"
	"net"

	"geodbsvc/internal/database"
)

// Lookup
func (d *SqliteDB) Lookup(ip string) (*database.GeoRecord, error) {
	netIP := net.ParseIP(ip)
	if netIP == nil {
		return nil, errors.New("wrong ip address format")
	}

	// Отсеиваем непубличные адреса
	if netIP.IsLoopback() || netIP.IsPrivate() || netIP.IsLinkLocalUnicast() || netIP.IsLinkLocalMulticast() {
		return nil, errors.New("ip is not a public address")
	}

	// Преобразуем к long
	netIPv4 := netIP.To4()
	if netIPv4 == nil {
		return nil, errors.New("ip is not an IPv4 address")
	}
	long := binary.BigEndian.Uint32(netIPv4)

	var res database.GeoRecord

	// Получаем запись из БД
	stmt, err := d.Prepare(`select
		continents.iso      "continent_code",
		continents.name_en  "continent_name_en",
		continents.name_ru  "continent_name_ru",
		countries.iso       "country_iso",
		countries.name_en   "country_name_en",
		countries.name_ru   "country_name_ru",
		regions.iso         "region_code",
		regions.name_en     "region_name_en",
		regions.name_ru     "region_name_ru",
		cities.name_en      "city_name_en",
		cities.name_ru      "city_name_ru",
		cities.lat          "city_lat",
		cities.lon          "city_lon",
		regions.timezone    "region_timezone"
	from networks
		left join countries on countries.id = networks.country_id
		left join continents on continents.id = countries.continent_id 
		left join regions on regions.id = networks.region_id
		left join cities on cities.id = networks.city_id
	where
		networks.ip <= :ip
	order by ip desc
	limit 1;`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var subdivision database.GeoSubdivision
	var regionCode sql.NullString
	var regionNameEn sql.NullString
	var regionNameRu sql.NullString
	var regionTimezone sql.NullString

	if err := stmt.QueryRow(sql.Named("ip", long)).Scan(
		&res.Continent.Code,     // continent_code
		&res.Continent.Names.En, // continent_name_en
		&res.Continent.Names.Ru, // continent_name_ru
		&res.Country.ISOCode,    // country_iso
		&res.Country.Names.En,   // country_name_en
		&res.Country.Names.Ru,   // country_name_ru
		&regionCode,             // region_code
		&regionNameEn,           // region_name_en
		&regionNameRu,           // region_name_ru
		&res.City.Names.En,      // city_name_en
		&res.City.Names.Ru,      // region_name_ru
		&res.Location.Latitude,  // city_lat
		&res.Location.Longitude, // city_lon
		&regionTimezone,         // region_timezone
	); err != nil {
		return nil, err
	}

	if regionCode.Valid {
		subdivision.ISOCode = regionCode.String
	}

	if regionNameEn.Valid {
		subdivision.Names.En = regionNameEn.String
	}

	if regionNameRu.Valid {
		subdivision.Names.Ru = regionNameRu.String
	}

	if regionTimezone.Valid {
		res.Location.TimeZone = regionTimezone.String
	}

	res.Subdivisions = append(res.Subdivisions, subdivision)

	return &res, nil
}
