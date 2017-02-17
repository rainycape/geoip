package geoip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultCacheDuration        = 24 * time.Hour
	minimumMaxMindCacheDuration = 24 * time.Hour
)

// GeoLiteKind indicates the kind of the GeoLite database.
// See the constants GeoLiteKindCountry and GeoLiteKindCity
// for more information.
type GeoLiteKind int

const (
	// GeoLiteKindCountry loads a database which contains only
	// country level records. It's faster and requires less
	// memory than GeoLiteKindCity.
	GeoLiteKindCountry GeoLiteKind = iota
	// GeoLiteKindCity loads a database with city level records.
	// Note that this is slower and consumes more memory than
	// GeoLiteKindCountry, so you should use only when you
	// need city level accuracy.
	GeoLiteKindCity
)

type urlOptions struct {
	CacheDir           string
	ExpirationDuration time.Duration
}

// URLOpt is a function type which allows setting options
// for opening geoip databases from HTTP(S) addresses
// directly. See URLCacheDir and URLCacheExpiration for
// more information.
type URLOpt func(*urlOptions)

// URLCacheDir sets the cache dir used for saving the Remote
// database locally. Note that, by default, the cache dir will
// be set to $HOME/.geoip. Use this function with an empty string
// argument to disable the cache.
func URLCacheDir(dir string) URLOpt {
	return func(opts *urlOptions) {
		opts.CacheDir = dir
	}
}

// URLCacheExpiration sets the cache expiration duration for the
// locally cached database. The default duration is 24 hours.
func URLCacheExpiration(duration time.Duration) URLOpt {
	return func(opts *urlOptions) {
		opts.ExpirationDuration = duration
	}
}

// OpenGeoLite opens a geoip2 database of the given kind from the
// MaxMind servers and caches it locally. See GeoLiteKind for the
// available database kinds. As for the available options, check
// OpenURL as the opts arguments it passed to it unmodified.
func OpenGeoLite(kind GeoLiteKind, opts ...URLOpt) (*GeoIP, error) {
	var url string
	switch kind {
	case GeoLiteKindCity:
		url = "http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.mmdb.gz"
	case GeoLiteKindCountry:
		url = "http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.mmdb.gz"
	default:
		return nil, fmt.Errorf("unknown GeoLite database kind %v", kind)
	}
	return OpenURL(url, opts...)
}

// OpenURL opens a geoip2 database from the given HTTP(S) URL. Use the functions
// URLCacheDir and URLCacheExpiration to your desired options. If the cache dir
// is not set, it will default to $HOME/.geoip. The default cache expiration time
// is 24 hours. Note that if you're loading the databases directly from MaxMind,
// this function will override expiration times lower than a day, to avoid overloading
// their servers.
func OpenURL(url string, opts ...URLOpt) (*GeoIP, error) {
	o := &urlOptions{
		ExpirationDuration: defaultCacheDuration,
	}
	if dir, err := defaultURLCacheDir(); err == nil {
		o.CacheDir = dir
	}
	for _, opt := range opts {
		opt(o)
	}
	duration := o.ExpirationDuration
	if strings.Contains(url, "maxmind.com") {
		// Avoid DDoS'ing MaxMind
		if duration < minimumMaxMindCacheDuration {
			duration = minimumMaxMindCacheDuration
		}
	}
	filename := filepath.Join(o.CacheDir, path.Base(url))
	st, err := os.Stat(filename)
	hasFile := err == nil
	if hasFile {
		now := time.Now()
		modTime := st.ModTime()
		expiration := modTime.Add(duration)
		// If modTime is in the future, assume something funny
		// happened and ignore the cached file for now.
		if modTime.Before(now) && expiration.After(now) {
			// The cached file exists and it's valid. Try to return it.
			// If it fails (e.g. the file got corrupted), fall back to
			// loading it from the URL.
			if db, err := Open(filename); err == nil {
				return db, nil
			}
		}
	}
	// The file doesn't exist or has expired
	db, data, err := openURL(url)
	if err != nil {
		// Remote loading failed. Try to fallback to
		// the cache.
		if hasFile {
			return Open(filename)
		}
		return nil, err
	}
	if o.CacheDir != "" {
		// Try to cache the data
		if err := os.MkdirAll(o.CacheDir, 0755); err == nil {
			if f, err := ioutil.TempFile(o.CacheDir, "geoip"); err == nil {
				if _, err := f.Write(data); err == nil {
					if err := f.Close(); err == nil {
						// Correctly cached into a temporary file, now
						// move to the cache path atomically.
						os.Rename(f.Name(), filename)
					}
				}
			}
		}
	}
	return db, nil
}

func defaultURLCacheDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		u, err := user.Current()
		if err != nil {
			// Can't recover from here, but a non
			// nil err should be rare enough that
			// it probably doesn't matter.
			return "", err
		}
		home = u.HomeDir
	}
	return filepath.Join(home, ".geoip"), nil
}

// open a *GeoIP from the given http(s) URL and return the
// raw body data too, so the caller can cache it.
func openURL(url string) (*GeoIP, []byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	decoded := data
	switch path.Ext(url) {
	case ".gz":
		gzr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, nil, err
		}
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, gzr); err != nil {
			return nil, nil, err
		}
		decoded = buf.Bytes()
	}
	db, err := New(bytes.NewReader(decoded))
	if err != nil {
		return nil, nil, err
	}
	return db, data, nil
}
