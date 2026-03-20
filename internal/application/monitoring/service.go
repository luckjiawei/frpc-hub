package monitoring

import (
	"podux/pkg/utils"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// GeoService resolves IP addresses to geolocation data with a 24-hour in-memory cache.
type GeoService struct {
	app       core.App
	cache     map[string]*geoCacheItem
	cacheLock sync.RWMutex
}

type geoCacheItem struct {
	Location  *utils.GeoLocation
	ExpiresAt time.Time
}

// NetworkStatus holds latency and reachability information.
type NetworkStatus struct {
	Latency       int64     `json:"latency"`
	Reachable     bool      `json:"reachable"`
	LastCheckTime time.Time `json:"lastCheckTime"`
	Error         string    `json:"error,omitempty"`
}

func NewGeoService(app core.App) *GeoService {
	return &GeoService{
		app:   app,
		cache: make(map[string]*geoCacheItem),
	}
}

// Lookup returns geolocation for the given address, using the cache when valid.
func (s *GeoService) Lookup(addr string) (*utils.GeoLocation, error) {
	s.cacheLock.RLock()
	if item, exists := s.cache[addr]; exists {
		if time.Now().Before(item.ExpiresAt) {
			s.cacheLock.RUnlock()
			return item.Location, nil
		}
	}
	s.cacheLock.RUnlock()

	location, err := utils.GetGeoLocation(addr)
	if err != nil {
		return nil, err
	}

	s.cacheLock.Lock()
	s.cache[addr] = &geoCacheItem{
		Location:  location,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	s.cacheLock.Unlock()

	return location, nil
}
