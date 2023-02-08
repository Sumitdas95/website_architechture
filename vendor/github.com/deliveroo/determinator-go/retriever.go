package determinator

import (
	"encoding/json"
)

// Retriever interface for implementing methods of pulling features.
type Retriever interface {
	Retrieve(name string) (Feature, error)
}

// ParseJSONFeature unmarshalls json into a Feature.
func ParseJSONFeature(raw []byte, buildMappedConstraints bool) (*FeatureData, error) {
	jsonFeature := FeatureData{}
	err := json.Unmarshal(raw, &jsonFeature)
	if err != nil {
		return &FeatureData{}, err
	}

	if buildMappedConstraints {
		jsonFeature.buildMappedConstraints()
	}

	return &jsonFeature, nil
}
