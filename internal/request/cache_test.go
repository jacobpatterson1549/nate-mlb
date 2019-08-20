package request

import (
	"strconv"
	"testing"
)

func TestNewCacheTooSmall(t *testing.T) {
	cacheSize := 0
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("wanted panic when trying to create cache with size=%d, but did not get one", cacheSize)
		}
	}()
	newCache(cacheSize)
}

func TestContainstNo(t *testing.T) {
	cache := newCache(5)
	url := "url"
	got := cache.contains(url)
	want := false
	if want != got {
		t.Errorf("wanted %v to not be in the cache, but it was", url)
	}
}

func TestContainsYes(t *testing.T) {
	cache := newCache(5)
	url := "url"
	cache.add(url, nil)
	got := cache.contains(url)
	want := true
	if want != got {
		t.Errorf("wanted %v to be in the cache, but it was not", url)
	}
}

func TestContainsNoAfterManyOtherAdds(t *testing.T) {
	cacheSize := 10
	cache := newCache(cacheSize)
	url := "url"
	cache.add(url, "abc")
	for i := 0; i < cacheSize; i++ {
		cache.add(strconv.Itoa(i), i)
	}
	got := cache.contains(url)
	want := false
	if want != got {
		t.Errorf("wanted %v to not be in the cache after it should be forgotten, but it was", url)
	}
}

func TestGet(t *testing.T) {
	cache := newCache(10)
	url := "url"
	value := "abc"
	cache.add(url, value)
	got := cache.get(url)
	want := value
	if want != got {
		t.Errorf("wanted %v too be in the cache for %v, but %v was present instead", want, url, got)
	}
}

func TestClear(t *testing.T) {
	cache := newCache(1)
	url := "url"
	cache.add(url, nil)
	cache.clear()
	got := cache.contains(url)
	want := false
	if want != got {
		t.Errorf("wanted cache to not contain url %v after clearing, but it was present", url)
	}
}
