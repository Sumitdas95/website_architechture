package determinator

import (
	"io/ioutil"
	"path/filepath"
)

// FileRetriever loads features from json files.
type FileRetriever struct {
	Root                   string
	buildMappedConstraints bool
}

// NewFileRetriever creates a new FileRetriever object.
func NewFileRetriever(path string, buildMappedConstraints bool) FileRetriever {
	retriever := FileRetriever{
		Root:                   path,
		buildMappedConstraints: buildMappedConstraints,
	}

	return retriever
}

// Retrieve reads the json file from the provided id.
func (retriever *FileRetriever) Retrieve(id string) (Feature, error) {
	path := filepath.Join(retriever.Root, id)

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	feature, err := ParseJSONFeature(raw, retriever.buildMappedConstraints)
	if err != nil {
		return nil, err
	}

	err = feature.validate()
	if err != nil {
		return nil, err
	}
	return feature, nil
}
