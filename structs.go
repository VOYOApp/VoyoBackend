package main

import (
	"encoding/json"
	"time"
)

//Omitempty permet de ne pas afficher le champs dans le json s'il est nil

type OTPData struct {
	PhoneNumber string `json:"phoneNumber" validate:"required"`
}

type VerifyData struct {
	User *OTPData `json:"user " validate:"required"`
	Code string   `json:"code " validate:"required"`
}

type TypeRealEstate struct {
	IdTypeRealEstate int       `json:"id"`
	Label            string    `json:"label"`
	Duration         time.Time `json:"duration"`
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
	PhoneNumber    string   `json:"phone_number"`
	FirstName      string   `json:"first_name"`
	LastName       string   `json:"last_name"`
	Email          string   `json:"email"`
	Password       string   `json:"password"`
	IdRole         int      `json:"role_id"`
	Biography      *string  `json:"biography"`
	ProfilePicture *string  `json:"profile_picture"`
	Pricing        *float64 `json:"pricing"`
	IdAddressGMap  *string  `json:"address_id"`
	Radius         *float64 `json:"radius"`
	X              *float64 `json:"x"`
	Y              *float64 `json:"y"`
	Geom           *string  `json:"geom"`
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var durationString string
	if err := json.Unmarshal(b, &durationString); err != nil {
		return err
	}

	// Parse the duration string
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		return err
	}

	d.Duration = duration
	return nil
}

type Availability struct {
	IdAvailability int       `json:"id"`
	PhoneNumber    string    `json:"phone_number"`
	Availability   time.Time `json:"availability"`
	Duration       Duration  `json:"duration"`
	Repeat         string    `json:"repeat"`
}

type Visit struct {
	IdVisit             int       `json:"id"`
	PhoneNumberProspect string    `json:"phone_number_prospect"`
	PhoneNumberVisitor  string    `json:"phone_number_visitor"`
	IdRealEstate        int       `json:"real_estate_id"`
	CodeVerification    int       `json:"verification_code"`
	StartTime           time.Time `json:"start_time"`
	Price               float64   `json:"price"`
	Status              string    `json:"status"`
	Note                float64   `json:"note"`
}

// Structs for unmarshaling the Google Maps API response
type googleMapsCoordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type googleMapsResponse struct {
	Results []struct {
		Geometry struct {
			Location googleMapsCoordinates `json:"location"`
		} `json:"geometry"`
	} `json:"results"`
	Status string `json:"status"`
}
