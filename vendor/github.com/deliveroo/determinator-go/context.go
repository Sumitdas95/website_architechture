package determinator

import (
	"context"
	"errors"
	"fmt"
)

type contextKey struct{}

// Context allows determinating on a set of actors.
type Context struct {
	Actors    Actors
	Retriever Retriever
	Collector TrackingCollector
}

// IsFeatureFlagOn determinates if the feature flag is on for the context.
func (c *Context) IsFeatureFlagOn(featureName string) (bool, error) {
	a, err := c.constructActor(featureName)
	if err != nil {
		return false, err
	}
	return a.feature.IsFeatureFlagOn(a.actor)
}

// WhichVariant checks which variant of a feature is active for context.
func (c *Context) WhichVariant(featureName string) (string, error) {
	a, err := c.constructActor(featureName)
	if err != nil {
		return "", err
	}
	return a.feature.WhichVariant(a.actor)
}

// CollectDeterminations returns the determinations from the configured TrackingCollector,
// if any.
func (c *Context) CollectDeterminations() map[Determination]int {
	if c.Collector == nil {
		return map[Determination]int{}
	}
	return c.Collector.CollectDeterminations()
}

// SetContext stores a determinator.Context into the context.Context.
func SetContext(ctx context.Context, dc *Context) context.Context {
	return context.WithValue(ctx, contextKey{}, dc)
}

// GetContext gets the determinator.Context from the context.Context.
func GetContext(ctx context.Context) (*Context, bool) {
	r, ok := ctx.Value(contextKey{}).(*Context)
	return r, ok
}

// NewContextFromCtx constructs a determinator.Context with the given Actors,
// using the Retriever and the Collector in ctx (if present).
func NewContextFromCtx(ctx context.Context, a Actors) (Context, error) {
	c := Context{Actors: a}
	r, ok := RetrieverFromContext(ctx)
	if !ok {
		return c, errors.New("Unable to get retriever from context")
	}
	c.Retriever = r
	collector, ok := TrackingCollectorFromContext(ctx)
	if ok {
		c.Collector = collector
	}
	return c, nil
}

type contextEvaluation struct {
	actor   Actor
	feature Feature
}

func (c *Context) constructActor(featureName string) (*contextEvaluation, error) {
	if c.Retriever == nil {
		return nil, errors.New("Retriever is not set")
	}
	feature, err := c.Retriever.Retrieve(featureName)
	if err != nil {
		return nil, err
	}
	featureData := feature.Data()
	if !featureData.IsStructured() {
		return nil, errors.New("Feature is not structured")
	}
	params, err := c.Actors.ToParams()
	if err != nil {
		return nil, err
	}
	val, ok := params[featureData.StructuredBucket]
	if !ok || val == "" {
		return nil, fmt.Errorf("%s not found in context", featureData.StructuredBucket)
	}
	actor := Actor{
		ID:         val,
		Properties: toProperties(params),
	}
	return &contextEvaluation{actor: actor, feature: feature}, nil
}

func toProperties(params map[string]string) map[string][]string {
	res := map[string][]string{}
	for k, v := range params {
		res[k] = []string{v}
	}
	return res
}
