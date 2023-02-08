package determinator

// MockRetriever implements the Retriever interface and is used for testing.
type MockRetriever struct {
	Features map[string]*FeatureData
	Validate bool
}

// NewMockRetriever returns a new instance of the MockRetriever.
func NewMockRetriever() MockRetriever {
	return MockRetriever{
		Features: map[string]*FeatureData{},
	}
}

// Retrieve loads a feature by name.
func (retriever *MockRetriever) Retrieve(featureID string) (Feature, error) {
	if feature, ok := retriever.Features[featureID]; ok {
		if retriever.Validate {
			err := feature.validate()
			if err != nil {
				return nil, err
			}
		}
		return feature, nil
	}
	return nil, nil
}

// Set sets a static feature in the MockRetriever.
func (retriever *MockRetriever) Set(feature *FeatureData) {
	retriever.Features[feature.ID] = feature
}
