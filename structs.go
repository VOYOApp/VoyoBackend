package main

import "time"

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type OTPData struct {
	PhoneNumber string `json:"phoneNumber,omitempty" validate:"required"`
}

type VerifyData struct {
	User *OTPData `json:"user,omitempty" validate:"required"`
	Code string   `json:"code,omitempty" validate:"required"`
}

type Bien struct {
	IdBien       int    `json:"id"`
	CodePostal   int    `json:"postal_code"`
	Ville        string `json:"city"`
	Adresse      string `json:"address"`
	Proprietaire string `json:"owner"`
	Pays         string `json:"country"`
}

type Role struct {
	IdRole  int    `json:"id"`
	Libelle string `json:"label"`
}

type Lieux struct {
	IdLieux    int    `json:"id"`
	Radius     string `json:"radius"`
	Adresse    string `json:"address"`
	Ville      string `json:"city"`
	CodePostal string `json:"postal_code"`
	Pays       string `json:"country"`
}

type Utilisateur struct {
	IdUtilisateur int     `json:"id"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	Adresse       string  `json:"address"`
	Ville         string  `json:"city"`
	CodePostal    int     `json:"postal_code"`
	Tel           string  `json:"tel"`
	Note          float64 `json:"note"`
	Description   string  `json:"description"`
	Password      string  `json:"password"`
	IdRole        int     `json:"role_id"`
}

type Calendrier struct {
	IdUtilisateur int       `json:"user_id"`
	IdCalendrier  int       `json:"calendar_id"`
	Disponibilite time.Time `json:"availability"`
	Temps         time.Time `json:"time"`
}

type Visite struct {
	IdUtilisateur    int       `json:"user_id"`
	IdUtilisateur1   int       `json:"user_id_1"`
	IdBien           int       `json:"property_id"`
	Agence           string    `json:"agency"`
	CodeVerification int       `json:"verification_code"`
	Horaire          time.Time `json:"schedule"`
	APayer           string    `json:"to_pay"`
	Etat             string    `json:"state"`
}

type Travail struct {
	IdUtilisateur int `json:"user_id"`
	IdLieux       int `json:"place_id"`
}
