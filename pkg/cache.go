package pkg

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

// Cache struct to manage the cache
type Cache struct {
	FileName string
	LifeTime time.Duration
	Expires  time.Time
}

var (
	// ErrCacheExpired is returned when cache expires
	ErrCacheExpired = errors.New("Cache expired")
	// CacheFolder is a folder to store cache files
	CacheFolder = "/tmp/aws_quota_exporter_cache/"
)

// NewCache creates a new Cache instance
func NewCache(fileName string, lifeTime time.Duration) *Cache {
	return &Cache{
		FileName: CacheFolder + fileName,
		LifeTime: time.Second * lifeTime,
		Expires:  time.Now(),
	}
}

// Read reads the contents of the cache file
func (c *Cache) Read() ([]*PrometheusMetric, error) {
	if time.Now().After(c.Expires) {
		return nil, ErrCacheExpired
	}
	byteData, err := os.ReadFile(c.FileName)
	if err != nil {
		return nil, err
	}
	var metrics []*PrometheusMetric
	err = json.Unmarshal(byteData, &metrics)

	return metrics, err
}

// Write writes data to the cache file
func (c *Cache) Write(data []*PrometheusMetric) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = os.WriteFile(c.FileName, jsonData, 0644)
	if err == nil {
		c.Expires = time.Now().Add(c.LifeTime)
	}
	return err
}
