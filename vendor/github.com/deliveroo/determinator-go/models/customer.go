package models

// Customer represents a person ordering on Deliveroo.
type Customer struct {
	ID       int `validate:"required"`
	DRN      *DRN
	Email    string `validate:"omitempty,email"`
	Employee bool
}
