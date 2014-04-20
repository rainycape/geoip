package geoip

import (
	"os"
	"testing"
)

func testNewGeoIP(t testing.TB) *GeoIP {
	const file = "GeoLite2-City.mmdb"
	geo, err := Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("missing file %s", file)
		}
		t.Fatal(err)
	}
	return geo
}

func TestGeoIP(t *testing.T) {
	geo := testNewGeoIP(t)
	rec1, err := geo.Lookup("1.2.3.4")
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("%+v", rec1)
	}
	rec2, err := geo.Lookup("17.0.0.1")
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("%+v", rec2)
	}
}

func TestOpenGz(t *testing.T) {
	const file = "GeoLite2-City.mmdb.gz"
	_, err := Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("missing file %s", file)
		}
		t.Fatal(err)
	}
}
