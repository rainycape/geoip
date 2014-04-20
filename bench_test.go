package geoip

import (
	"net"
	"testing"
)

func BenchmarkLookup(b *testing.B) {
	geo := testNewGeoIP(b)
	ip := net.ParseIP("17.0.0.1")
	b.ResetTimer()
	for ii := 0; ii < b.N; ii++ {
		geo.LookupIP(ip)
	}
}
