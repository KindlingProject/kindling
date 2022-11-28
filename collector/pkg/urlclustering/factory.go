package urlclustering

func NewMethod(urlClusteringMethod string) ClusteringMethod {
	switch urlClusteringMethod {
	case "alphabet":
		return NewAlphabeticalClusteringMethod()
	case "noparam":
		return NewNoParamClusteringMethod()
	case "blank":
		return NewBlankClusteringMethod()
	default:
		return NewAlphabeticalClusteringMethod()
	}
}
