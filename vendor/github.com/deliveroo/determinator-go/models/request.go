package models

type RequestCustomer struct {
	GUID        string `validate:"required"`
	SessionGUID string
	StickyGUID  string
}

type Request struct {
	UID             string `validate:"required"`
	Country         *Country
	City            *City
	Zone            *Zone
	Platform        string `validate:"oneof=web ios android bot api signature-api"`
	WhiteLabelBrand string
	AppVersion      string
	Customer        *RequestCustomer
}
