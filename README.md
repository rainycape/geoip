geoip
=====

[![GoDoc](https://godoc.org/github.com/rainycape/geoip?status.svg)](https://godoc.org/github.com/rainycape/geoip)

GeoIP2 library in Go (golang)

This library implements reading and decoding of GeoIP2 databases. Free
databases can be downloaded from [MaxMind][1]. 

To install geoip run the following command:

```
    go get github.com/rainycape/geoip
```

## Example

```go
package main

import (
	"fmt"
	"github.com/rainycape/vfs"
)

func main() {
	db, err := geoip.Open("GeoLite2-City.mmdb.gz")
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
```

## License

This code is licensed under the [MPL 2.0][2].

[1]: http://dev.maxmind.com/geoip/geoip2/geolite2/
[2]: http://www.mozilla.org/MPL/2.0/
