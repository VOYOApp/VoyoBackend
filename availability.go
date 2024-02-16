package main

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
)

// CreateAvailability crée une nouvelle disponibilité dans la base de données.
func CreateAvailability(c *fiber.Ctx) error {
	var availability Availability
	if err := c.BodyParser(&availability); err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO availability (PhoneNumber, Availability, Duration, Repeat) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(availability.PhoneNumber, availability.Availability, availability.Duration, availability.Repeat)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).SendString("Disponibilité créée avec succès")
}

// GetAvailability récupère une disponibilité spécifique à partir de son ID.
func GetAvailability(c *fiber.Ctx) error {
	id := c.Query("id")

	var availability Availability
	err := db.QueryRow("SELECT PhoneNumber, Availability, Duration, Repeat FROM availability WHERE ID = $1", id).Scan(&availability.PhoneNumber, &availability.Availability, &availability.Duration, &availability.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Disponibilité non trouvée")
		}
		return err
	}

	return c.JSON(availability)
}

// UpdateAvailability met à jour une disponibilité existante dans la base de données.
func UpdateAvailability(c *fiber.Ctx) error {
	id := c.Query("id")

	var availability Availability
	if err := c.BodyParser(&availability); err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE availability SET PhoneNumber=$1, Availability=$2, Duration=$3, Repeat=$4 WHERE ID=$5")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(availability.PhoneNumber, availability.Availability, availability.Duration, availability.Repeat, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteAvailability supprime une disponibilité de la base de données.
func DeleteAvailability(c *fiber.Ctx) error {
	id := c.Query("id")

	stmt, err := db.Prepare("DELETE FROM availability WHERE ID=$1")
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
