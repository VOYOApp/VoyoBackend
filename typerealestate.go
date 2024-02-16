package main

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
)

// CreateTypeRealEstate crée un nouveau type de bien immobilier dans la base de données.
func CreateTypeRealEstate(c *fiber.Ctx) error {
	var typeRealEstate TypeRealEstate
	if err := c.BodyParser(&typeRealEstate); err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO typerealestate (Label, Duration) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(typeRealEstate.Label, typeRealEstate.Duration)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).SendString("Type de bien immobilier créé avec succès")
}

// GetTypeRealEstate récupère un type de bien immobilier spécifique à partir de son ID.
func GetTypeRealEstate(c *fiber.Ctx) error {
	id := c.Query("id")

	var typeRealEstate TypeRealEstate
	err := db.QueryRow("SELECT IdTypeRealEstate, Label, Duration FROM typerealestate WHERE IdTypeRealEstate = $1", id).Scan(&typeRealEstate.IdTypeRealEstate, &typeRealEstate.Label, &typeRealEstate.Duration)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Type de bien immobilier non trouvé")
		}
		return err
	}

	return c.JSON(typeRealEstate)
}

// UpdateTypeRealEstate met à jour un type de bien immobilier existant dans la base de données.
func UpdateTypeRealEstate(c *fiber.Ctx) error {
	id := c.Query("id")

	var typeRealEstate TypeRealEstate
	if err := c.BodyParser(&typeRealEstate); err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE typerealestate SET Label=$1, Duration=$2 WHERE IdTypeRealEstate=$3")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(typeRealEstate.Label, typeRealEstate.Duration, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteTypeRealEstate supprime un type de bien immobilier de la base de données.
func DeleteTypeRealEstate(c *fiber.Ctx) error {
	id := c.Query("id")

	stmt, err := db.Prepare("DELETE FROM typerealestate WHERE IdTypeRealEstate=$1")
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
