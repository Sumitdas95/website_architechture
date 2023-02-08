package determinator

import "sync"

// TrackingCollector represents a struct which returns a list of determinations.
type TrackingCollector interface {
	CollectDeterminations() map[Determination]int
}

// SingleTrackingCollector is a simple container which holds a TrackingRetriever. It is useful
// when the TrackingRetriever is not available when setting up the tracking, and is instead created
// later in the request lifecycle.
type SingleTrackingCollector struct {
	retriever *TrackingRetriever
	mu        sync.Mutex
}

// SetRetriever sets the retriever instance.
func (tc *SingleTrackingCollector) SetRetriever(r *TrackingRetriever) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.retriever = r
}

// CollectDeterminations returns the retriever's determinations.
func (tc *SingleTrackingCollector) CollectDeterminations() map[Determination]int {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.retriever == nil {
		return map[Determination]int{}
	}
	return tc.retriever.CollectDeterminations()
}

// MockCollector returns a fixed list of determinations
type MockCollector struct {
	Determinations map[Determination]int
}

// CollectDeterminations conforms to the TrackingCollector interface.
func (m *MockCollector) CollectDeterminations() map[Determination]int {
	return m.Determinations
}
