package request

import (
	"fmt"
	"sync"
)

type cache struct {
	requestValues map[string][]byte
	requestURIs   []string
	index         int
	mutex         *sync.Mutex
}

func newCache(cacheSize int) cache {
	if cacheSize < 0 {
		panic(fmt.Sprintf("cache size must be positive - got %v", cacheSize))
	}
	return cache{
		requestValues: make(map[string][]byte, cacheSize),
		requestURIs:   make([]string, cacheSize),
		index:         0,
		mutex:         &sync.Mutex{},
	}
}

func (c *cache) contains(uri string) bool {
	_, ok := c.requestValues[uri]
	return ok
}

func (c *cache) add(uri string, value []byte) {
	if c.index >= len(c.requestURIs) {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.requestValues, c.requestURIs[c.index])
	c.requestValues[uri] = value
	c.requestURIs[c.index] = uri
	c.index++
	if c.index == len(c.requestURIs) {
		c.index = 0
	}
}

func (c *cache) get(uri string) ([]byte, bool) {
	b, ok := c.requestValues[uri]
	return b, ok
}

func (c *cache) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for k := range c.requestValues {
		delete(c.requestValues, k)
	}
	for i := range c.requestURIs {
		c.requestURIs[i] = ""
	}
	c.index = 0
}
