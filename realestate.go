package main

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
)

// CreateRealEstate crée un nouveau bien immobilier dans la base de données.
func CreateRealEstate(c *fiber.Ctx) error {
	var realEstate RealEstate
	if err := c.BodyParser(&realEstate); err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO realestate (IdAddressGMap, IdTypeRealEstate) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(realEstate.IdAddressGMap, realEstate.IdTypeRealEstate)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).SendString("Bien immobilier créé avec succès")
}

// GetRealEstate récupère un bien immobilier spécifique à partir de son ID.
func GetRealEstate(c *fiber.Ctx) error {
	id := c.Query("id")

	var realEstate RealEstate
	err := db.QueryRow("SELECT IdRealEstate, IdAddressGMap, IdTypeRealEstate FROM realestate WHERE IdRealEstate = $1", id).Scan(&realEstate.IdRealEstate, &realEstate.IdAddressGMap, &realEstate.IdTypeRealEstate)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Bien immobilier non trouvé")
		}
		return err
	}

	return c.JSON(realEstate)
}

// UpdateRealEstate met à jour un bien immobilier existant dans la base de données.
func UpdateRealEstate(c *fiber.Ctx) error {
	id := c.Query("id")

	var realEstate RealEstate
	if err := c.BodyParser(&realEstate); err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE realestate SET IdAddressGMap=$1, IdTypeRealEstate=$2 WHERE IdRealEstate=$3")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(realEstate.IdAddressGMap, realEstate.IdTypeRealEstate, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteRealEstate supprime un bien immobilier de la base de données.
func DeleteRealEstate(c *fiber.Ctx) error {
	id := c.Query("id")

	stmt, err := db.Prepare("DELETE FROM realestate WHERE IdRealEstate=$1")
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
