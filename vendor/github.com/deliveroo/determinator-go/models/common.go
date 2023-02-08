package models

type DRN struct {
	ID     string `validate:"required"`
	Market string `validate:"required"`
}

type City struct {
	ID    int `validate:"required"`
	DRN   *DRN
	UName string `param:"uname" validate:"required"`
}

type Country struct {
	ID  int `validate:"required"`
	DRN *DRN
	TLD string `validate:"required"`
}

type Neighborhood struct {
	ID  int `validate:"required"`
	DRN *DRN
}

type Zone struct {
	ID   int `validate:"required"`
	DRN  *DRN
	Code string `validate:"required"`
}
