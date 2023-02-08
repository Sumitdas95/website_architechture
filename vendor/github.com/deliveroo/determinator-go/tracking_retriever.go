package determinator

import (
	"sync"
)

// TrackingRetriever implements the Retriever interface and stores the results of the Determinations.
type TrackingRetriever struct {
	detInstance     Retriever
	featureTrackers []*FeatureTracker
	mux             sync.Mutex
}

// Retrieve returns a Feature and any associated errors.
func (tr *TrackingRetriever) Retrieve(featureID string) (Feature, error) {
	res, err := tr.detInstance.Retrieve(featureID)
	if err != nil {
		return res, err
	}

	if res == nil {
		return nil, nil
	}

	trackingRes := CreateFeatureTracker(res)
	tr.mux.Lock()
	tr.featureTrackers = append(tr.featureTrackers, trackingRes)
	tr.mux.Unlock()
	return trackingRes, nil
}

// CollectDeterminations gets all the Determinations from the FeatureTrackers.
func (tr *TrackingRetriever) CollectDeterminations() map[Determination]int {
	tr.mux.Lock()
	defer tr.mux.Unlock()

	// Pre-size the map to the total number of feature trackers. Reduce heap
	// allocation to the already known target size.
	ds := make(map[Determination]int, len(tr.featureTrackers))

	fts := tr.featureTrackers
	for _, f := range fts {
		for k, v := range f.GetDeterminations() {
			ds[k] += v
		}
	}
	return ds
}

// ConstructTrackingRetriever takes a Retriever and wraps it in a TrackingRetriever.
func ConstructTrackingRetriever(detInstance Retriever) TrackingRetriever {
	return TrackingRetriever{
		detInstance: detInstance,
		mux:         sync.Mutex{},
	}
}
