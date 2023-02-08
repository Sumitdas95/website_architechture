package determinator

import (
	"sync"
)

// FeatureTracker contains a instance of FeatureData and stores all the Determinations made.
type FeatureTracker struct {
	determinations map[Determination]int
	feature        Feature
	mux            sync.Mutex
}

// CreateFeatureTracker initializes a FeatureTracker.
func CreateFeatureTracker(feature Feature) *FeatureTracker {
	return &FeatureTracker{
		determinations: make(map[Determination]int),
		feature:        feature,
		mux:            sync.Mutex{},
	}
}

// GetDeterminations returns the Determinations.
func (t *FeatureTracker) GetDeterminations() map[Determination]int {
	t.mux.Lock()
	defer t.mux.Unlock()
	return t.determinations
}

// Data returns the data for the feature.
func (t *FeatureTracker) Data() *FeatureData {
	return t.feature.Data()
}

// IsExperiment calls the wrapped FeatureData's IsExperiment.
func (t *FeatureTracker) IsExperiment() bool {
	return t.feature.IsExperiment()
}

// IsFeatureFlagOn calls IsFeatureFlagOn and keeps track of the Determinations made.
func (t *FeatureTracker) IsFeatureFlagOn(actor Actor) (bool, error) {
	inc, err := t.feature.IsFeatureFlagOn(actor)
	t.noticeDetermination(actor.ID, actor.GUID, t.feature.Data(), DeterminationResponse{Include: inc})
	return inc, err
}

// WhichVariant calls WhichVariant and keeps track of the Determinations made.
func (t *FeatureTracker) WhichVariant(actor Actor) (string, error) {
	v, err := t.feature.WhichVariant(actor)
	t.noticeDetermination(actor.ID, actor.GUID, t.feature.Data(), DeterminationResponse{Variant: v})
	return v, err
}

// Determinate calls Determinate and keeps track of the Determinations made.
func (t *FeatureTracker) Determinate(actor Actor) DeterminationResponse {
	dr := t.feature.Determinate(actor)
	t.noticeDetermination(actor.ID, actor.GUID, t.feature.Data(), dr)
	return dr
}

func (t *FeatureTracker) noticeDetermination(id string, guid string, feature *FeatureData, dr DeterminationResponse) {
	t.mux.Lock()
	defer t.mux.Unlock()
	if feature == nil {
		return
	}
	featureID := feature.Identifier
	if featureID == "" {
		featureID = feature.ID
	}
	t.determinations[Determination{
		ID:        id,
		GUID:      guid,
		FeatureID: featureID,
		DeterminationResponse: DeterminationResponse{
			Include: dr.Include,
			Variant: dr.Variant,
		},
	}]++
}
