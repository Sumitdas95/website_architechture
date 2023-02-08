# 0.5.4

- `tracking_retriever.go` now pre-sizes the `CollectDeterminations` map to reduce heap allocations

# 0.5.3

- Remove errors from tracked DeterminationResponse

# 0.5.2

- `http_retriever.go` now returns an error message when the response is a 404.
  - `errors.Wrapf` has now been updated to use `fmt.Errorf` as the errors package is deprecated.
  - Switch from `ioutils.ReadAll` to `io.ReadAll` as `ioutils.ReadAll` is deprecated.

# 0.5.1

- Performance improvements for flags with many constraint values
  - Constraint values are now stored as maps, alongside the existing array values.
  - Depending on how your client code is caching features, you may need to republish
    features being consumed by your app to benefit from this change.

# 0.5.0

- `feature-codegen` generated code can now log errors that occur when speaking to Determinator

# 0.4.1

- Fix parsing protobuf features (ensure no slices or maps are nil)

# 0.4.0

- Added a command-line utility, `feature-codegen` to remove boilerplate by generating code to interact with feature flags

# 0.3.0

Breaking change:
- `TrackingRetriever` now returns `nil, nil` (a nil feature and no error) when the feature requested does not exist (bringing it inline with other retrievers).
  - Previously it would have returned a `FeatureTracker` wrapping a nil `Feature` which would have led to a nil pointer panic if any of the functions were called on it
- `MockRetriever` now returns `nil, nil` (a nil feature and no error) when the feature requested was not added to the mock (bringing it inline with how production retrievers behave)
  - When upgrading to this version it is likely to make tests fail if they were expecting errors to be returned for non-existent features. However, correcting these tests will ensure they are asserting on the true production behaviour

# 0.2.1

- Fix percentage rollout, when not using an actor or guid, so that it is random

# 0.2.0

- Fix structured bucket

# 0.1.7

- Upgrade NewFeatureFromProtobufMessage

# 0.1.6

- Make Country, City and Zone optional in request actor

# 0.1.5

- Add Collector to DeterminatorContext
- Add bot and api platforms to actors

# 0.1.4

- Use go 1.13
- Support request.app_version

# 0.1.3

- Updated transport-models module to support latest Feature protobuf definition

# 0.1.2

- Changed FeatureTracker to accept the Feature interface instead of FeatureData, to help if mocking is needed.

# 0.1.1

- `WhichVariant` will not throw an error anymore if the tested flag is not experiment and inactive.
- `IsFeatureFlagOn` will not throw an error anymore if the tested flag is an experiment and inacive.

Feature:

- Add fixed determinations

# 0.1.0

Breaking change:

- Feature has been changed from a struct to a interface

Feature:

- Allow users to track feature determinations and provide middleware for tracking
