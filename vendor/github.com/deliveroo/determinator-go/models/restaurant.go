package models

type Restaurant struct {
	ID           int `validate:"required"`
	DRN          *DRN
	Country      Country
	City         City
	Neighborhood Neighborhood
	Zone         Zone
}
