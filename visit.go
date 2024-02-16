package main

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
)

// CreateVisit crée une nouvelle visite dans la base de données.
func CreateVisit(c *fiber.Ctx) error {
	var visit Visit
	if err := c.BodyParser(&visit); err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO visit (PhoneNumberProspect, PhoneNumberVisitor, IdRealEstate, CodeVerification, StartTime, Price, Status, Note) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(visit.PhoneNumberProspect, visit.PhoneNumberVisitor, visit.IdRealEstate, visit.CodeVerification, visit.StartTime, visit.Price, visit.Status, visit.Note)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).SendString("Visite créée avec succès")
}

// GetVisit récupère une visite spécifique à partir de son ID.
func GetVisit(c *fiber.Ctx) error {
	id := c.Query("id")

	var visit Visit
	err := db.QueryRow("SELECT PhoneNumberProspect, PhoneNumberVisitor, IdRealEstate, CodeVerification, StartTime, Price, Status, Note FROM visit WHERE ID = $1", id).Scan(&visit.PhoneNumberProspect, &visit.PhoneNumberVisitor, &visit.IdRealEstate, &visit.CodeVerification, &visit.StartTime, &visit.Price, &visit.Status, &visit.Note)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Visite non trouvée")
		}
		return err
	}

	return c.JSON(visit)
}

// UpdateVisit met à jour une visite existante dans la base de données.
func UpdateVisit(c *fiber.Ctx) error {
	id := c.Query("id")

	var visit Visit
	if err := c.BodyParser(&visit); err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE visit SET PhoneNumberProspect=$1, PhoneNumberVisitor=$2, IdRealEstate=$3, CodeVerification=$4, StartTime=$5, Price=$6, Status=$7, Note=$8 WHERE ID=$9")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(visit.PhoneNumberProspect, visit.PhoneNumberVisitor, visit.IdRealEstate, visit.CodeVerification, visit.StartTime, visit.Price, visit.Status, visit.Note, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteVisit supprime une visite de la base de données.
func DeleteVisit(c *fiber.Ctx) error {
	id := c.Query("id")

	stmt, err := db.Prepare("DELETE FROM visit WHERE ID=$1")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
