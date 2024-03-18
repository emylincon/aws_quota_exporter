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
func NewCache(fileName string, lifeTime time.Duration) (*Cache, error) {
	// check if cache folder exists
	if _, err := os.Stat(CacheFolder); os.IsNotExist(err) {
		err = os.MkdirAll(CacheFolder, 0755)
		if err != nil {
			return nil, errors.New("Error creating cache folder: " + err.Error())
		}
	}

	f, err := os.CreateTemp(CacheFolder, fileName+"-*.json")
	if err != nil {
		return nil, errors.New("Could not initialise cache for " + fileName + ": " + err.Error())
	}

	return &Cache{
		FileName: f.Name(),
		LifeTime: time.Second * lifeTime,
		Expires:  time.Now(),
	}, nil
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
