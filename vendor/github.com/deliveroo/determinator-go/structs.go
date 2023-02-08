package determinator

import "time"

// Actor defines user which will be the basis for the indicator for determination.
type Actor struct {
	ID         string              `json:"id,omitempty"`
	GUID       string              `json:"guid,omitempty"`
	Properties map[string][]string `json:"properties,omitempty"`
	Location   string              `json:"location,omitempty"`
}

// DeterminationResponse defines the response for the determinator.
type DeterminationResponse struct {
	Include bool   `json:"include"`
	Variant string `json:"variant,omitempty"`
	Error   error  `json:"error,omitempty"`
}

// Determination defines what gets stored by the FeatureTracker.
type Determination struct {
	ID                    string
	GUID                  string
	FeatureID             string
	DeterminationResponse DeterminationResponse
}

// DeterminationContext stores data for the request along with all the Determinations for that request.
type DeterminationContext struct {
	RequestType    string
	Path           string
	Started        time.Time
	Finished       time.Time
	Status         int
	Determinations map[Determination]int
}

// DeterminationError defines the error response for the determinator.
type DeterminationError struct {
	Message string
	Fatal   bool
}

func (err DeterminationError) Error() string {
	return err.Message
}
