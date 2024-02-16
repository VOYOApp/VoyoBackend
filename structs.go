package main

import (
	"time"
)

//omitempty permet de ne pas afficher le champs dans le json s'il est nil

type OTPData struct {
	PhoneNumber string `json:"phoneNumber,omitempty" validate:"required"`
}

type VerifyData struct {
	User *OTPData `json:"user,omitempty" validate:"required"`
	Code string   `json:"code,omitempty" validate:"required"`
}

type TypeRealEstate struct {
	IdTypeRealEstate int           `json:"id"`
	Label            string        `json:"label"`
	Duration         time.Duration `json:"duration"`
}

type RealEstate struct {
	IdRealEstate     int    `json:"id"`
	IdAddressGMap    string `json:"address_id"`
	IdTypeRealEstate int    `json:"type_id"`
}

type Role struct {
	IdRole int    `json:"id"`
	Label  string `json:"label"`
}

type User struct {
	PhoneNumber    string  `json:"phone_number,omitempty"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Email          string  `json:"email,omitempty"`
	Description    string  `json:"description"`
	Password       string  `json:"password"`
	IdRole         int     `json:"role_id"`
	Biography      string  `json:"biography,omitempty"`
	ProfilePicture string  `json:"profile_picture,omitempty"`
	Pricing        float64 `json:"pricing,omitempty"`
	IdAddressGMap  string  `json:"address_id,omitempty"`
	Radius         string  `json:"radius,omitempty"`
}

type Availability struct {
	IdAvailability int       `json:"id"`
	PhoneNumber    string    `json:"phone_number"`
	Availability   time.Time `json:"availability"`
	Duration       time.Time `json:"duration"`
	Repeat         string    `json:"repeat"`
}

type Visit struct {
	PhoneNumberProspect string    `json:"phone_number_prospect"`
	PhoneNumberVisitor  string    `json:"phone_number_visitor"`
	IdRealEstate        int       `json:"real_estate_id"`
	CodeVerification    int16     `json:"verification_code"`
	StartTime           time.Time `json:"start_time"`
	Price               string    `json:"price"`
	Status              string    `json:"status"`
	Note                float64   `json:"note,omitempty"`
}
