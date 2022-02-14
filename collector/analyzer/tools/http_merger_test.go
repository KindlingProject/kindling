package tools

import (
	"testing"
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
