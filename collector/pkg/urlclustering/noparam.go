package urlclustering

import (
	"strings"
)

// NoParamClusteringMethod removes the parameters that are end of the URL string.
type NoParamClusteringMethod struct {
}

func NewNoParamClusteringMethod() ClusteringMethod {
	return &NoParamClusteringMethod{}
}

func (m *NoParamClusteringMethod) Clustering(endpoint string) string {
	if endpoint == "" {
		return ""
	}
	// Remove the parameters first
	index := strings.Index(endpoint, "?")
	if index != -1 {
		endpoint = endpoint[:index]
	}

	return endpoint
}
