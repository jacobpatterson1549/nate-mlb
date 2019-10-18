package request

import (
	"reflect"
	"strconv"
	"testing"
)

func TestNewCacheTooSmall(t *testing.T) {
	cacheSize := -1
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("wanted panic when trying to create cache with size=%d, but did not get one", cacheSize)
		}
	}()
	NewCache(cacheSize)
}

func TestContainsNo(t *testing.T) {
	cache := NewCache(5)
	uri := "uri"
	got := cache.contains(uri)
	want := false
	if want != got {
		t.Errorf("wanted %v to not be in the cache, but it was", uri)
	}
}

func TestContainsZero(t *testing.T) {
	cache := NewCache(0)
	uri := "uri"
	cache.add(uri, nil)
	got := cache.contains(uri)
	want := false
	if want != got {
		t.Errorf("wanted %v to not be in the cache, but it was", uri)
	}
}

func TestContainsYes(t *testing.T) {
	cache := NewCache(5)
	uri := "uri"
	cache.add(uri, nil)
	got := cache.contains(uri)
	want := true
	if want != got {
		t.Errorf("wanted %v to be in the cache, but it was not", uri)
	}
}

func TestContainsNoAfterManyOtherAdds(t *testing.T) {
	cacheSize := 10
	cache := NewCache(cacheSize)
	uri := "uri"
	cache.add(uri, nil)
	for i := 0; i < cacheSize; i++ {
		j := strconv.Itoa(i)
		cache.add(j, []byte(j))
	}
	got := cache.contains(uri)
	want := false
	if want != got {
		t.Errorf("wanted %v to not be in the cache after it should be forgotten, but it was", uri)
	}
}

func TestGet(t *testing.T) {
	cache := NewCache(10)
	uri := "uri"
	value := []byte("abc")
	cache.add(uri, value)
	got, ok := cache.get(uri)
	want := value
	switch {
	case !ok:
		t.Error("wanted value to be present, but it was not")
	case !reflect.DeepEqual(want, got):
		t.Errorf("wanted\n%v to be in the cache for %v, but\n%v was present instead", string(want), uri, string(got))
	}
}

func TestClear(t *testing.T) {
	cache := NewCache(1)
	uri := "uri"
	cache.add(uri, nil)
	cache.Clear()
	got := cache.contains(uri)
	want := false
	if want != got {
		t.Errorf("wanted cache to not contain uri %v after clearing, but it was present", uri)
	}
}
