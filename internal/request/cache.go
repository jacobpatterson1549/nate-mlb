package request

import "fmt"

type cache struct {
	requestValues map[string]interface{}
	requestUrls   []string
	index         int
}

func newCache(cacheSize int) cache {
	if cacheSize <= 0 {
		panic(fmt.Sprintf("cache size must be positive - got %v", cacheSize))
	}
	return cache{
		requestValues: make(map[string]interface{}),
		requestUrls:   make([]string, cacheSize),
		index:         0,
	}
}

func (c *cache) contains(url string) bool {
	_, ok := c.requestValues[url]
	return ok
}

func (c *cache) add(url string, value interface{}) {
	delete(c.requestValues, c.requestUrls[c.index])
	c.requestValues[url] = value
	c.requestUrls[c.index] = url
	c.index++
	if c.index == len(c.requestUrls) {
		c.index = 0
	}
}

func (c *cache) get(url string) interface{} {
	return c.requestValues[url]
}

func (c *cache) clear() {
	for k := range c.requestValues {
		delete(c.requestValues, k)
	}
	for i := range c.requestUrls {
		c.requestUrls[i] = ""
	}
	c.index = 0
}
