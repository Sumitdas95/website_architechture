package determinator

import (
	"context"
	"net/http"
)

type trackingCollectorKey struct{}
type retrieverKey struct{}

// ContextWithTrackingCollector stores a TrackingCollector into the context.
func ContextWithTrackingCollector(ctx context.Context, dc TrackingCollector) context.Context {
	return context.WithValue(ctx, trackingCollectorKey{}, dc)
}

// ContextWithRetriever stores a Retriever into the context.
func ContextWithRetriever(ctx context.Context, dc Retriever) context.Context {
	return context.WithValue(ctx, retrieverKey{}, dc)
}

// RetrieverFromContext gets the TrackingRetriever from the context.
func RetrieverFromContext(ctx context.Context) (Retriever, bool) {
	r, ok := ctx.Value(retrieverKey{}).(Retriever)
	return r, ok
}

// TrackingCollectorFromContext gets the TrackingCollector from the context.
func TrackingCollectorFromContext(ctx context.Context) (TrackingCollector, bool) {
	c, ok := ctx.Value(trackingCollectorKey{}).(TrackingCollector)
	return c, ok
}

// Determinating provides a middleware to wrap the Retriever in.
func Determinating(retriever Retriever) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			createDeterminationTrackerAndServe(retriever, h, w, r)
		})
	}
}

func createDeterminationTrackerAndServe(retriever Retriever, h http.Handler, w http.ResponseWriter, r *http.Request) {
	dt := ConstructTrackingRetriever(retriever)
	ctx := ContextWithRetriever(r.Context(), &dt)
	ctx = ContextWithTrackingCollector(ctx, &dt)

	r = r.WithContext(ctx)
	h.ServeHTTP(w, r)
}
