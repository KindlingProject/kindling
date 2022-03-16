package tools

import (
	"regexp"
	"strings"
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
)

const (
	KEY_THRESHOLD = 5
)

var (
	alphabetRegexp *regexp.Regexp
	starRegexp     *regexp.Regexp
)

func init() {
	alphabetRegexp, _ = regexp.Compile("^[A-Za-z_-]+$")
	starRegexp, _ = regexp.Compile("^[/*]+$")
}

type HttpMerger struct {
	keyThreshold int
	cache        *simplelru.LRU
	mutex        sync.RWMutex
}

func NewHttpMergeCache() *HttpMerger {
	merger := &HttpMerger{}
	merger.keyThreshold = KEY_THRESHOLD
	merger.cache, _ = simplelru.NewLRU(20000, nil)
	return merger
}

func (merger *HttpMerger) GetContentKey(url string) string {
	if len(url) <= 1 {
		return url
	}

	// If the url is not starting with '/', return "*"
	if url[0] != '/' {
		return "*"
	}

	algorithmConvergen := merger.convergeByAlgorithm(url)
	regexConvergence := merger.convergeByRegex(algorithmConvergen)
	return truncateStarUrl(regexConvergence)
}

func (merger *HttpMerger) convergeByAlgorithm(url string) string {
	urlAt := getSubUrl(url, 0)
	if len(urlAt) == 0 {
		// If url_at is empty, the url is "/"
		return url
	}

	urlBt := getSubUrl(url, len(urlAt))
	if len(urlBt) == 0 {
		// If url_bt is empty, the url is like "/foo"
		return url
	}

	urlAt = urlAt + "/*"
	urlBt = "/*" + urlBt
	// If the input parameter is "/*/*", the return value is either empty or "/*/*" itself
	result := merger.getCacheKey(urlAt, urlBt)
	if len(result) == 0 {
		return url
	}
	return result
}

func (merger *HttpMerger) getCacheKey(urlA string, urlB string) string {
	if len(urlA) == 0 || len(urlB) == 0 {
		return ""
	}

	valueA, valueB := merger.getCountByUrl(urlA, urlB)

	if valueA < 0 && valueB > 0 {
		merger.addUrlCount(urlA, 1, urlB, valueB+1)
		if valueB+1 >= merger.keyThreshold {
			return urlB
		}
	} else if valueA > 0 && valueB < 0 {
		merger.addUrlCount(urlA, valueA+1, urlB, 1)
		if valueA+1 >= merger.keyThreshold {
			return urlA
		}
	} else if valueA > 0 && valueB > 0 {
		var res = -1
		var resStr = ""
		if valueA > valueB {
			res = valueA
			resStr = urlA
		} else {
			res = valueB
			resStr = urlB
		}
		if res >= merger.keyThreshold {
			return resStr
		} else {
			return ""
		}
	} else {
		merger.addUrlCount(urlA, 1, urlB, 1)
		return ""
	}

	return ""
}

func (merger *HttpMerger) getCountByUrl(urlA string, urlB string) (int, int) {
	merger.mutex.Lock()
	defer merger.mutex.Unlock()

	valueA := -1
	if cache, ok := merger.cache.Get(urlA); ok {
		valueA = cache.(int)
	}
	valueB := -1
	if cache, ok := merger.cache.Get(urlB); ok {
		valueB = cache.(int)
	}
	return valueA, valueB
}

func (merger *HttpMerger) addUrlCount(urlA string, countA int, urlB string, countB int) {
	merger.mutex.Lock()
	defer merger.mutex.Unlock()

	merger.cache.Add(urlA, countA)
	merger.cache.Add(urlB, countB)
}

func (merger *HttpMerger) convergeByRegex(url string) string {
	// "" -> ""
	// "aa" -> "aa"
	// "/" -> "", ""
	// "/aa" -> "", "aa"
	// "/aa/" -> "", "aa", ""
	// "/aa/bb" -> "", "aa", "bb"
	// "/aa/bb/" -> "", "aa", "bb", ""
	segments := strings.Split(url, "/")
	// 1. first part
	if len(segments) < 2 {
		// Means that there is no "/"
		return url
	}

	var firstPart = segments[1]
	if len(firstPart) == 0 {
		// The url is "/"
		return url
	}
	match := alphabetRegexp.MatchString(firstPart)
	if !match {
		firstPart = "*"
	}

	var ret = "/" + firstPart
	// 2. second part
	if len(segments) < 3 {
		return ret
	}
	var secondPart = segments[2]
	if len(secondPart) == 0 {
		return ret + "/"
	}
	if !isUrlAlphabet(secondPart) {
		secondPart = "*"
	}
	ret += "/" + secondPart
	return ret
}

func (merger *HttpMerger) setKeyThreshold(value int) {
	merger.keyThreshold = value
}

func isUrlAlphabet(s string) bool {
	match, _ := regexp.MatchString("^[A-Za-z]+$", s)
	return match
}

func getSubUrl(url string, posi int) string {
	if len(url) == 0 || (posi >= len(url)-1) || posi < 0 {
		return ""
	}

	index := strings.Index(url[posi+1:], "/")
	if index == -1 {
		return url[posi:]
	}

	if index <= 0 {
		return ""
	}

	return url[posi : index+posi+1]
}

func truncateStarUrl(url string) string {
	if len(url) == 0 {
		return url
	}
	match := starRegexp.MatchString(url)
	if match {
		return "*"
	}
	return url
}
