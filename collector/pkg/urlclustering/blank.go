package urlclustering

// BlankClusteringMethod removes the endpoint and return an empty string.
// This method is used to reduce the cardinality as much as possible.
type BlankClusteringMethod struct {
}

func NewBlankClusteringMethod() ClusteringMethod {
	return &BlankClusteringMethod{}
}

func (m *BlankClusteringMethod) Clustering(_ string) string {
	return ""
}
