package geoip_test

import (
	"fmt"

	"github.com/rainycape/geoip"
)

func ExampleOpen() {
	db, err := geoip.OpenGeoLite(geoip.GeoLiteKindCity)
	if err != nil {
		panic(err)
	}
	res, err := db.Lookup("17.0.0.1")
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Country.Name)
	fmt.Println(res.City.Name)
	// Output:
	// United States
	// Cupertino
}
