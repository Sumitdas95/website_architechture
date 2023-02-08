package models

type Rider struct {
	ID      int    `validate:"required"`
	UUID    string `validate:"required"`
	DRN     *DRN
	Zone    Zone
	Country Country
}
