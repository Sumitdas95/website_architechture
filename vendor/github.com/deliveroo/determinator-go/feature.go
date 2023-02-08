package determinator

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/blang/semver"
	transport "github.com/deliveroo/transport-models-go/types/messages/platform/actor_tracking"
)

// FixedDetermination defines constraints that will fix a user's determination.
type FixedDetermination struct {
	Name              string              `json:"name"`
	FeatureOn         bool                `json:"feature_on"`
	Variant           string              `json:"variant"`
	Constraints       map[string][]string `json:"constraints"`
	MappedConstraints Constraints         `json:"mapped_constraints"`
}

// TargetGroup defines constraints of the feature it belongs to.
type TargetGroup struct {
	Name              string              `json:"name"`
	Rollout           uint32              `json:"rollout"`
	Constraints       map[string][]string `json:"constraints"`
	MappedConstraints Constraints         `json:"mapped_constraints"`
}

// Constraints is a type alias for storing allowed constraint values in maps, for speedier lookups
type Constraints map[string]ConstraintValues

// ConstraintValues is a type alias for storing allowed values for a single constraint
type ConstraintValues map[string]bool

// Feature is the interface for implementing methods for checking information about a feature and checking how a actor interacts with the feature.
type Feature interface {
	Data() *FeatureData
	IsExperiment() bool
	IsFeatureFlagOn(actor Actor) (bool, error)
	WhichVariant(actor Actor) (string, error)
	Determinate(actor Actor) DeterminationResponse
}

// FeatureData defines the feature, its constraints, variants and active status.
type FeatureData struct {
	ID                  string               `json:"id"`
	Name                string               `json:"name"`
	Identifier          string               `json:"identifier"`
	BucketType          string               `json:"bucket_type"`
	StructuredBucket    string               `json:"structured_bucket"`
	FixedDeterminations []FixedDetermination `json:"fixed_determinations"`
	TargetGroups        []TargetGroup        `json:"target_groups"`
	Variants            map[string]uint16    `json:"variants"`
	WinningVariant      string               `json:"winning_variant"`
	Active              bool                 `json:"active"`
	Overrides           map[string]string    `json:"overrides"`
}

// Data returns FeatureData.
func (f *FeatureData) Data() *FeatureData {
	return f
}

// IsExperiment determines if this is an A/B test or just a feature flag.
func (f *FeatureData) IsExperiment() bool {
	return len(f.Variants) != 0
}

// IsStructured returns true if the feature is using a structured bucket.
func (f *FeatureData) IsStructured() bool {
	return f.StructuredBucket != ""
}

func (f *FeatureData) validate() error {
	for _, fixedDetermination := range f.FixedDeterminations {
		if !fixedDetermination.FeatureOn && fixedDetermination.Variant != "" {
			return errors.New("feature_on must be set to true if a Variant is specified")
		}

		_, ok := f.Variants[fixedDetermination.Variant]
		if fixedDetermination.Variant != "" && !ok {
			return fmt.Errorf("Variant %s is not contained in the variants", fixedDetermination.Variant)
		}
	}

	for _, targetGroup := range f.TargetGroups {
		if targetGroup.Rollout > 65536 {
			return errors.New("feature has invalid rollout")
		}
	}

	return nil
}

// IsFeatureFlagOn determinates if the feature flag is on for a actor.
func (f *FeatureData) IsFeatureFlagOn(actor Actor) (bool, error) {
	if !f.Active {
		return false, nil
	}
	if f.IsExperiment() {
		return false, errors.New("feature is not a feature flag")
	}
	response := f.Determinate(actor)
	return response.Include, response.Error
}

// WhichVariant checks which variant a actor belongs to.
func (f *FeatureData) WhichVariant(actor Actor) (string, error) {
	if !f.Active {
		return "", nil
	}
	if !f.IsExperiment() {
		return "", errors.New("feature is not an experiment")
	}
	response := f.Determinate(actor)
	return response.Variant, response.Error
}

// Determinate will determine which variant a user is in based on the DeterminationRequest.
func (f *FeatureData) Determinate(actor Actor) DeterminationResponse {
	response := DeterminationResponse{Include: false}

	if !f.Active {
		return response
	}

	// check if the determination has been fixed
	fixedDeterminationResponse := chooseFixedDetermination(f.FixedDeterminations, actor.Properties)
	if fixedDeterminationResponse != nil {
		return *fixedDeterminationResponse
	}

	rollout := chooseRollout(f.TargetGroups, actor.Properties)
	if rollout == 0 {
		return response
	}

	rolloutIndicator, variantIndicator, err := indicators(f, actor.ID, actor.GUID)
	if err != nil {
		response.Error = err
		return response
	}
	if rolloutIndicator >= rollout {
		return response
	}

	response.Include = true

	if !f.IsExperiment() {
		return response
	}

	response.Variant = chooseVariant(f, variantIndicator)
	return response
}

// Enriches the FeatureData struct with mappings of constraint values
// If present, these will be used in preference to the default value arrays when doing inclusion checks
func (f *FeatureData) buildMappedConstraints() {
	for i, df := range f.FixedDeterminations {
		if df.MappedConstraints != nil {
			continue
		}

		mapped := make(Constraints)

		for name, values := range df.Constraints {
			mapped[name] = newConstraintValues(values)
		}

		f.FixedDeterminations[i].MappedConstraints = mapped
	}
	for i, tg := range f.TargetGroups {
		if tg.MappedConstraints != nil {
			continue
		}

		mapped := make(Constraints)

		for name, values := range tg.Constraints {
			mapped[name] = newConstraintValues(values)
		}

		f.TargetGroups[i].MappedConstraints = mapped
	}
}

func chooseFixedDetermination(fixedDeterminations []FixedDetermination, properties map[string][]string) *DeterminationResponse {
	for _, fixedDetermination := range fixedDeterminations {

		fixedDeterminationMatches := true

		for requiredKey, values := range fixedDetermination.Constraints {
			givenValues, ok := properties[requiredKey]
			if !ok {
				fixedDeterminationMatches = false
				break
			}

			//We want to match over all the constraints in a fixed determination
			var mappedValues = fixedDetermination.MappedConstraints.getMappedConstraintValuesIfPresent(requiredKey)

			fixedDeterminationMatches = matchesConstraints(requiredKey, values, mappedValues, givenValues)
			if !fixedDeterminationMatches {
				break
			}
		}

		if fixedDeterminationMatches {
			return &DeterminationResponse{
				Include: fixedDetermination.FeatureOn,
				Variant: fixedDetermination.Variant,
			}
		}
	}
	return nil
}

func chooseRollout(targetGroups []TargetGroup, properties map[string][]string) uint32 {
	if len(targetGroups) == 0 {
		return 0
	}

	var matchingRollout uint32

	for _, targetGroup := range targetGroups {
		targetGroupMatches := true

		for requiredKey, allowedValues := range targetGroup.Constraints {
			givenValues, ok := properties[requiredKey]
			if !ok {
				targetGroupMatches = false
				break
			}

			//We want to match over all the constraints in a target group
			var mappedValues = targetGroup.MappedConstraints.getMappedConstraintValuesIfPresent(requiredKey)

			targetGroupMatches = matchesConstraints(requiredKey, allowedValues, mappedValues, givenValues)
			if !targetGroupMatches {
				break
			}
		}

		if targetGroupMatches {
			if targetGroup.Rollout == 65536 {
				// Can't get any higher, may as well return now
				return 65536
			}
			if targetGroup.Rollout > matchingRollout {
				matchingRollout = targetGroup.Rollout
			}
		}
	}

	return matchingRollout
}

func matchesConstraints(key string, allowedValues []string, mappedValues ConstraintValues, givenValues []string) bool {
	switch key {
	case "app_version", "request.app_version":
		return matchesAppVersionConstraints(allowedValues, givenValues)
	default:
		return matchesEqualConstraints(allowedValues, mappedValues, givenValues)
	}
}

func matchesAppVersionConstraints(allowedValues []string, givenValues []string) bool {
	expectedRangeString := strings.Join(allowedValues, " ")
	expectedRange, err := semver.ParseRange(expectedRangeString)
	if err != nil {
		return false
	}
	for _, givenValue := range givenValues {
		v, err := semver.Parse(givenValue)
		if err != nil {
			return false
		}
		if expectedRange(v) {
			return true
		}
	}
	return false
}

func matchesEqualConstraints(allowedValues []string, mappedValues ConstraintValues, givenValues []string) bool {
	if mappedValues != nil {
		// Use map lookup for faster presence check, if mapped values are available
		for _, givenValue := range givenValues {
			if mappedValues[givenValue] {
				return true
			}
		}
	} else {
		// Fall back to checking each array value
		for _, allowedValue := range allowedValues {
			for _, givenValue := range givenValues {
				if givenValue == allowedValue {
					return true
				}
			}
		}
	}

	return false
}

func indicators(feature *FeatureData, id string, guid string) (uint32, uint16, error) {
	if feature == nil {
		return 0, 0, errors.New("feature cannot be nil")
	}
	indicatorString := feature.Identifier + ","

	switch feature.BucketType {
	case "guid":
		if len(guid) == 0 {
			return 65536, 0, errors.New("a GUID must be given for GUID bucketed features")
		}
		indicatorString = indicatorString + guid
	case "id":
		if len(id) == 0 {
			return 65536, 0, nil
		}
		indicatorString = indicatorString + id
	case "fallback":
		if len(id) == 0 {
			if len(guid) == 0 {
				return 0, 0, errors.New("an ID or GUID must be given for Fallback bucketed features")
			}
			indicatorString = indicatorString + guid
		} else {
			indicatorString = indicatorString + id
		}
	case "single":
		var err error
		indicatorString, err = randomHexString(64)
		if err != nil {
			return 65536, 0, fmt.Errorf("generating random value for single indicator: %w", err)
		}
	default:
		return 65536, 0, errors.New("unknown bucket type")
	}

	hash := md5.Sum([]byte(indicatorString))

	rollout := uint32(binary.BigEndian.Uint16(hash[0:2]))
	variant := binary.BigEndian.Uint16(hash[2:4])

	return rollout, variant, nil
}

func chooseVariant(feature *FeatureData, variantIndicator uint16) string {
	if feature == nil {
		return ""
	}
	if feature.WinningVariant != "" {
		return feature.WinningVariant
	}

	weightTotal := 0.0
	variants := make([]string, len(feature.Variants))
	varIdx := 0
	for variant, weight := range feature.Variants {
		weightTotal = weightTotal + float64(weight)
		variants[varIdx] = variant
		varIdx = varIdx + 1
	}
	scaleFactor := 65535 / weightTotal

	sort.Strings(variants)

	upperBound := 0.0
	for _, variant := range variants {
		upperBound = upperBound + (float64(feature.Variants[variant]) * scaleFactor)
		if variantIndicator <= uint16(upperBound) {
			return variant
		}
	}

	// We get here when the float division in generating scaleFactor means
	// the final total of upperBound is just less than 65535, just pick the
	// last variant, as that would have been the right one
	return variants[varIdx-1]
}

// NewFeatureFromProtobufMessage transform a protobuf message of type "Feature"
// into a determinator.Feature
func NewFeatureFromProtobufMessage(feature *transport.Feature) FeatureData {
	variants := make(map[string]uint16, len(feature.Variants))
	for _, v := range feature.Variants {
		variants[v.Name] = uint16(v.Value)
	}

	fixedDeterminations := make([]FixedDetermination, len(feature.FixedDeterminations))
	for i, fd := range feature.FixedDeterminations {
		v := FixedDetermination{
			Name:        fd.Name,
			FeatureOn:   fd.FeatureOn,
			Variant:     fd.Variant,
			Constraints: make(map[string][]string, len(fd.Constraints)),
		}

		for _, c := range fd.Constraints {
			if c.Values == nil {
				// Ensure nil values are serialised to JSON as empty array
				v.Constraints[c.Key] = []string{}
			} else {
				v.Constraints[c.Key] = c.Values
			}
		}

		fixedDeterminations[i] = v
	}

	targetGroups := make([]TargetGroup, len(feature.TargetGroups))
	for i, tg := range feature.TargetGroups {
		v := TargetGroup{
			Name:        tg.Name,
			Rollout:     tg.Rollout,
			Constraints: make(map[string][]string, len(tg.Constraints)),
		}

		for _, c := range tg.Constraints {
			if c.Values == nil {
				// Ensure nil values are serialised to JSON as empty array
				v.Constraints[c.Key] = []string{}
			} else {
				v.Constraints[c.Key] = c.Values
			}
		}

		targetGroups[i] = v
	}

	f := FeatureData{
		ID:                  feature.Id,
		Name:                feature.Name,
		Identifier:          feature.Identifier,
		BucketType:          feature.BucketType,
		FixedDeterminations: fixedDeterminations,
		TargetGroups:        targetGroups,
		Variants:            variants,
		WinningVariant:      feature.WinningVariant,
		Active:              feature.Active,
	}

	f.buildMappedConstraints()

	return f
}

func randomHexString(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func newConstraintValues(values []string) ConstraintValues {
	mapped := make(ConstraintValues)

	for _, v := range values {
		mapped[v] = true
	}

	return mapped
}

func (c *Constraints) getMappedConstraintValuesIfPresent(name string) ConstraintValues {
	if c == nil {
		return nil
	}

	return (*c)[name]
}
