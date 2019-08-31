package request

import (
	"fmt"
	"sync"
)

type cache struct {
	requestValues map[string][]byte
	requestUrls   []string
	index         int
	mutex         *sync.Mutex
}

func newCache(cacheSize int) cache {
	if cacheSize <= 0 {
		panic(fmt.Sprintf("cache size must be positive - got %v", cacheSize))
	}
	return cache{
		requestValues: make(map[string][]byte, cacheSize),
		requestUrls:   make([]string, cacheSize),
		index:         0,
		mutex:         &sync.Mutex{},
	}
}

func (c *cache) contains(url string) bool {
	_, ok := c.requestValues[url]
	return ok
}

func (c *cache) add(url string, value []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.requestValues, c.requestUrls[c.index])
	c.requestValues[url] = value
	c.requestUrls[c.index] = url
	c.index++
	if c.index == len(c.requestUrls) {
		c.index = 0
	}
}

func (c *cache) get(url string) ([]byte, bool) {
	b, ok := c.requestValues[url]
	return b, ok
}

func (c *cache) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for k := range c.requestValues {
		delete(c.requestValues, k)
	}
	for i := range c.requestUrls {
		c.requestUrls[i] = ""
	}
	c.index = 0
}
