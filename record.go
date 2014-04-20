package geoip

import (
	"fmt"
)

var (
	codes = []string{"iso_code", "code"}
)

type Name map[string]string

func (n Name) String() string {
	return n["en"]
}

func (n Name) LocalizedName(lang string) string {
	return n[lang]
}

func (n Name) Localizations() []string {
	keys := make([]string, 0, len(n))
	for k := range n {
		keys = append(keys, k)
	}
	return keys
}

type Place struct {
	Code      string
	GeonameID int
	Name      Name
}

func (p *Place) String() string {
	return p.Name.String()
}

type Record struct {
	Continent           *Place
	Country             *Place
	RegisteredCountry   *Place
	RepresentedCountry  *Place
	City                *Place
	Subdivisions        []*Place
	Latitude            float64
	Longitude           float64
	MetroCode           string
	PostalCode          string
	TimeZone            string
	IsAnonymousProxy    bool
	IsSatelliteProvider bool
}

func (r *Record) CountryCode() string {
	if r != nil && r.Country != nil {
		return r.Country.Code
	}
	return ""
}

func newPlace(val interface{}) *Place {
	if m, ok := val.(map[string]interface{}); ok {
		geonameId := int(m["geoname_id"].(uint32))
		var code string
		for _, v := range codes {
			if c, ok := m[v].(string); ok {
				code = c
				break
			}
		}
		name := make(Name)
		if names, ok := m["names"].(map[string]interface{}); ok {
			for k, v := range names {
				if n, ok := v.(string); ok {
					name[k] = n
				}
			}
		}
		return &Place{
			Code:      code,
			GeonameID: geonameId,
			Name:      name,
		}
	}
	return nil
}

func newRecord(val interface{}) (*Record, error) {
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid record type %T", val)
	}
	var latitude, longitude float64
	var metroCode, postalCode, timeZone string
	var isAnonymousProxy, isSatelliteProvider bool
	if location, ok := m["location"].(map[string]interface{}); ok {
		latitude, _ = location["latitude"].(float64)
		longitude, _ = location["longitude"].(float64)
		if m := location["metro_code"]; m != nil {
			metroCode = fmt.Sprintf("%v", m)
		}
		timeZone, _ = location["time_zone"].(string)
	}
	if postal, ok := m["postal"].(map[string]interface{}); ok {
		postalCode, _ = postal["code"].(string)
	}
	var subdivisions []*Place
	if subs, ok := m["subdivisions"].([]interface{}); ok {
		for _, v := range subs {
			if p := newPlace(v); p != nil {
				subdivisions = append(subdivisions, p)
			}
		}
	}
	if traits, ok := m["traits"].(map[string]interface{}); ok {
		isAnonymousProxy, _ = traits["is_anonymous_proxy"].(bool)
		isSatelliteProvider, _ = traits["is_satellite_provider"].(bool)
	}
	return &Record{
		Continent:           newPlace(m["continent"]),
		Country:             newPlace(m["country"]),
		RegisteredCountry:   newPlace(m["registered_country"]),
		RepresentedCountry:  newPlace(m["represented_country"]),
		City:                newPlace(m["city"]),
		Subdivisions:        subdivisions,
		Latitude:            latitude,
		Longitude:           longitude,
		MetroCode:           metroCode,
		PostalCode:          postalCode,
		TimeZone:            timeZone,
		IsAnonymousProxy:    isAnonymousProxy,
		IsSatelliteProvider: isSatelliteProvider,
	}, nil
}
