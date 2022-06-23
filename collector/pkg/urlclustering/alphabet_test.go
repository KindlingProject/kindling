package urlclustering

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	endpoint string
	want     string
}

func newTestcases() []testCase {
	return []testCase{
		{"/", "/"},
		{"/A/b/20000/", "/A/b/*/"},
		{" /test/22?v=a", "/test/*"},
		{"/a12?a=1132", "/*"},
		{"/abcd/1234a/efg/b&*", "/abcd/*/efg/*"},
		// Double slashes is valid but not recommended
		{"/a//b/c?d=2&e=3", "/a//b/c"},
	}
}

func TestAlphabeticClusteringMethod_Clustering(t *testing.T) {
	method := NewAlphabeticalClusteringMethod()
	testCases := newTestcases()
	for _, c := range testCases {
		assert.Equal(t, c.want, method.Clustering(c.endpoint))
	}
}

func TestAlphabeticClusteringMethod_ClusteringBaseline(t *testing.T) {
	method := NewAlphabeticalClusteringMethod()
	testCases := newTestcases()
	for _, c := range testCases {
		assert.Equal(t, c.want, method.ClusteringBaseline(c.endpoint))
	}
}

func Test_isAlphabetical(t *testing.T) {
	type args struct {
		b byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"alphabetical", args{'a'}, true},
		{"non-alphabetical", args{'%'}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAlphabetical(tt.args.b); got != tt.want {
				t.Errorf("isAlphabetical() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Benchmark_ClusteringBaseline(b *testing.B) {
	method := NewAlphabeticalClusteringMethod()
	testCases := newTestcases()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, c := range testCases {
			method.ClusteringBaseline(c.endpoint)
		}
	}
}

func Benchmark_Clustering(b *testing.B) {
	method := NewAlphabeticalClusteringMethod()
	testCases := newTestcases()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, c := range testCases {
			method.Clustering(c.endpoint)
		}
	}
}
