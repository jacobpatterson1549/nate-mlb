package request

import (
	"fmt"
	"sync"
)

// Cache keeps track of recent requests.
// A queue is used to track recent requests.  The stored recent request is returned as long as the request uri is in the cache.
type Cache struct {
	requestValues map[string][]byte
	requestURIs   []string
	index         int
	mutex         *sync.Mutex
}

// NewCache creates a Cache with a specified queue size
func NewCache(queueSize int) Cache {
	if queueSize < 0 {
		panic(fmt.Sprintf("cache size must be positive - got %v", queueSize))
	}
	return Cache{
		requestValues: make(map[string][]byte, queueSize),
		requestURIs:   make([]string, queueSize),
		index:         0,
		mutex:         &sync.Mutex{},
	}
}

func (c *Cache) contains(uri string) bool {
	_, ok := c.requestValues[uri]
	return ok
}

func (c *Cache) add(uri string, value []byte) {
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

func (c *Cache) get(uri string) ([]byte, bool) {
	b, ok := c.requestValues[uri]
	return b, ok
}

// Clear removes stored responses for all the URIs is the Cache
func (c *Cache) Clear() {
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
