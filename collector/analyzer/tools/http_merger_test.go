package tools

import (
	"strconv"
	"sync"
	"testing"

	"github.com/hashicorp/golang-lru/simplelru"
)

func TestHttpMerger_ConvergeByAlgorithm(t *testing.T) {
	httpMerger := NewHttpMergeCache()
	httpMerger.setKeyThreshold(3)

	tests := []struct {
		url  string
		want string
	}{
		{"/A/1", "/A/1"},
		{"/A/1", "/A/1"},
		{"/A/1", "/A/1"},
		{"/A/5", "/A/5"},
		{"/1/B", "/1/B"},
		{"/5/B", "/5/B"},
		{"/A/B", "/A/B"},
		{"/A/C", "/A/*"},
		{"/A/D", "/A/*"},
		{"/C/B", "/*/B"},
		{"/Cddd", "/Cddd"},
		{"/C/", "/C/"},
		{"/C/", "/C/"},
		{"/C/", "/C/"},
		{"/C/", "/C/"},
		{"asdd", "asdd"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := httpMerger.convergeByAlgorithm(tt.url); got != tt.want {
				t.Errorf("convergeByAlgorithm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpMerger_ConvergeByRegex(t *testing.T) {
	httpMerger := NewHttpMergeCache()
	httpMerger.setKeyThreshold(3)

	tests := []struct {
		url  string
		want string
	}{
		{"/", "/"},
		{"/", "/"},
		{"/a", "/a"},
		{"/a/", "/a/"},
		{"/a/b", "/a/b"},
		{"/a/*", "/a/*"},
		{"/a-c/*", "/a-c/*"},
		{"/a-c1/*", "/*/*"},
		{"/a-c/b-d", "/a-c/*"},
		{"/*/b-d", "/*/*"},
		{"/kc-main-commodity/api/", "/kc-main-commodity/api"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := httpMerger.convergeByRegex(tt.url); got != tt.want {
				t.Errorf("convergeByRegex() got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpMerger_TruncateStarUrl(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"*", "*"},
		{"/*", "*"},
		{"/*/", "*"},
		{"/*/*", "*"},
		{"/*/*/", "*"},
		{"/a/*", "/a/*"},
		{"/a/b", "/a/b"},
		{"/*/b", "/*/b"},
		{"/a*/*", "/a*/*"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := truncateStarUrl(tt.url); got != tt.want {
				t.Errorf("convergeByRegex() got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpMerger_Cache(t *testing.T) {
	httpMerger := &HttpMerger{}
	httpMerger.cache, _ = simplelru.NewLRU(3, nil)

	tests := []struct {
		testName string
		size     int
		urlA     string
		urlB     string
		wantA    int
		wantB    int
	}{
		{"Empty Cache", 0, "a", "b", -1, -1},      // Empty
		{"Match One Key", 2, "a", "c", 1, -1},     // a, b
		{"Match Another Key", 3, "a", "d", 2, -1}, // b, a, c
		{"Overflow", 3, "c", "d", 1, 1},           // c, a, d
		{"LRU Cache", 3, "a", "b", 3, -1},         // a, c, d
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			cacheSize := httpMerger.cache.Len()
			if cacheSize != tt.size {
				t.Errorf("Expect cache Size got %v, want %v", cacheSize, tt.size)
			}
			valueA, valueB := httpMerger.getCountByUrl(tt.urlA, tt.urlB)
			if valueA != tt.wantA {
				t.Errorf("Expect valueA got %v, want %v", valueA, tt.wantA)
			}
			if valueB != tt.wantB {
				t.Errorf("Expect valueB got %v, want %v", valueB, tt.wantB)
			}

			if valueA < 0 && valueB > 0 {
				httpMerger.addUrlCount(tt.urlA, 1, tt.urlB, valueB+1)
			} else if valueA > 0 && valueB < 0 {
				httpMerger.addUrlCount(tt.urlA, valueA+1, tt.urlB, 1)
			} else if valueA < 0 && valueB < 0 {
				httpMerger.addUrlCount(tt.urlA, 1, tt.urlB, 1)
			}
		})
	}
}

func TestConcurrentLruCache(t *testing.T) {
	httpMerger := &HttpMerger{}
	httpMerger.cache, _ = simplelru.NewLRU(10, nil)
	var wg sync.WaitGroup

	addFunc := func() {
		defer func() {
			err := recover()
			if err != nil {
				t.Errorf("Error Occured %v", err)
			}
			wg.Done()
		}()

		urlA, urlB := "aaa", "bbb"
		for i := 0; i < 100; i++ {
			suffix := strconv.Itoa(i)
			httpMerger.getCountByUrl(urlA+suffix, urlB+suffix)
			httpMerger.addUrlCount(urlA+suffix, 1, urlB+suffix, 1)
		}

		wg.Done()
	}

	for threads := 0; threads < 4; threads++ {
		wg.Add(1)
		go addFunc()
	}

	wg.Wait()
}
