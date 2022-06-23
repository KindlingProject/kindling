package urlclustering

import (
	"regexp"
	"strings"
)

// AlphabeticClusteringMethod clustering all the non-alphabetic characters to '*'
// and will trim the parameters at the end of the string.
type AlphabeticClusteringMethod struct {
	regexp *regexp.Regexp
}

func NewAlphabeticalClusteringMethod() *AlphabeticClusteringMethod {
	exp, _ := regexp.Compile("^[A-Za-z_-]+$")
	return &AlphabeticClusteringMethod{
		regexp: exp,
	}
}

func (m *AlphabeticClusteringMethod) ClusteringBaseline(endpoint string) string {
	if endpoint == "" {
		return ""
	}
	index := strings.Index(endpoint, "?")
	if index != -1 {
		endpoint = endpoint[:index]
	}

	// Split the endpoint into multiple segments
	endpoint = strings.TrimSpace(endpoint)
	segments := strings.Split(endpoint, "/")
	// Iterate over all parts and execute the regular expression.
	resultSegments := make([]string, 0, len(segments))
	// Skip the first segment because it is supposed to be always empty.
	for i := 1; i < len(segments); i++ {
		if segments[i] == "" || m.regexp.MatchString(segments[i]) {
			resultSegments = append(resultSegments, segments[i])
		} else {
			// If the segment is composed of non-alphabetic characters, we replace it with a star.
			resultSegments = append(resultSegments, "*")
		}
	}
	// Re-combine all parts together
	var resultEndpoint string
	for _, seg := range resultSegments {
		resultEndpoint = resultEndpoint + "/" + seg
	}

	return resultEndpoint
}

func (m *AlphabeticClusteringMethod) Clustering(endpoint string) string {
	if endpoint == "" {
		return ""
	}
	endpointBytes := []byte(endpoint)

	resultBytes := make([]byte, 0)

	currentSegmentIsStar := false
	currentSegment := make([]byte, 0)

	for _, b := range endpointBytes {
		if b == ' ' {
			continue
		}
		// End of the current segment.
		if b == '/' || b == '?' {
			if currentSegmentIsStar {
				resultBytes = append(resultBytes, '*')
			} else {
				// currentSegment could be empty
				resultBytes = append(resultBytes, currentSegment...)
			}
			currentSegment = make([]byte, 0)
			currentSegmentIsStar = false
			if b == '/' {
				resultBytes = append(resultBytes, '/')
				continue
			} else if b == '?' {
				break
			}
		}
		if currentSegmentIsStar {
			continue
		}
		if isAlphabetical(b) {
			currentSegment = append(currentSegment, b)
		} else {
			currentSegmentIsStar = true
		}
	}
	// Deal with the last segment.
	if currentSegmentIsStar {
		resultBytes = append(resultBytes, '*')
	} else if len(currentSegment) != 0 {
		resultBytes = append(resultBytes, currentSegment...)
	}
	return string(resultBytes)
}

func isAlphabetical(b byte) bool {
	return b == '-' || b == '_' || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
