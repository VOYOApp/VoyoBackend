package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"time"
)

// CreateAvailability crée une nouvelle disponibilité dans la base de données.
func CreateAvailability(c *fiber.Ctx) error {
	var availability []Availability
	if err := c.BodyParser(&availability); err != nil {
		fmt.Println("💥 Error parsing the body in CreateAvailability() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	for _, a := range availability {
		// Convert duration to seconds
		// Chaîne de caractères représentant l'heure
		heureString := "04:29:00"

		// Convertir la chaîne en type de données Time
		heureTime, err := time.Parse("15:04:05", heureString)
		if err != nil {
		}

		a.PhoneNumber = c.Locals("user").(*CustomClaims).PhoneNumber

		stmt, err := db.Prepare("INSERT INTO availability (PhoneNumber, Availability, Duration, Repeat) VALUES ($1, $2, $3::interval, $4)")
		if err != nil {
			fmt.Println("💥 Error preparing the SQL statement in CreateAvailability() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		defer func(stmt *sql.Stmt) {
			err := stmt.Close()
			if err != nil {
				fmt.Println("💥 Error closing the SQL statement in CreateAvailability() : ", err)
				return
			}
		}(stmt)

		_, err = stmt.Exec(a.PhoneNumber, a.Availability, heureTime.Format("15:04:05"), a.Repeat)
		if err != nil {
			fmt.Println("💥 Error executing the SQL statement in CreateAvailability() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
	}

	return c.Status(fiber.StatusCreated).SendString("Disponibilité créée avec succès")
}

// GetAvailability récupère une disponibilité spécifique à partir de son ID, ou toutes les disponibilités s'il n'y a pas d'ID spécifié.
func GetAvailability(c *fiber.Ctx) error {
	id := c.Query("id")

	// Si un ID est spécifié dans les paramètres de la requête,
	// on récupère uniquement cette disponibilité spécifique.
	if id != "" {
		var availability Availability
		stmt, err := db.Prepare("SELECT IdAvailability, PhoneNumber, Availability, Duration, Repeat FROM availability WHERE IdAvailability = $1")
		if err != nil {
			fmt.Println("💥 Error preparing the SQL statement in GetAvailability() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		availability.PhoneNumber = c.Locals("user").(User).PhoneNumber

		defer func(stmt *sql.Stmt) {
			err := stmt.Close()
			if err != nil {
				fmt.Println("💥 Error closing the SQL statement in GetAvailability() : ", err)
				return
			}
		}(stmt)

		row := stmt.QueryRow(id)
		err = row.Scan(&availability.IdAvailability, &availability.PhoneNumber, &availability.Availability, &availability.Duration, &availability.Repeat)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).SendString("Availability not found")
			}

			fmt.Println("💥 Error scanning the row in GetAvailability() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		return c.JSON(availability)
	}

	// Si aucun ID n'est spécifié, on récupère toutes les disponibilités.
	rows, err := db.Query("SELECT IdAvailability, PhoneNumber, Availability, Duration, Repeat FROM availability")
	if err != nil {
		fmt.Println("💥 Error querying the database in GetAvailability() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println("💥 Error closing the rows in GetAvailability() : ", err)
			return
		}
	}(rows)

	var availabilities []Availability
	for rows.Next() {
		var availability Availability
		err := rows.Scan(&availability.IdAvailability, &availability.PhoneNumber, &availability.Availability, &availability.Duration, &availability.Repeat)
		if err != nil {
			fmt.Println("💥 Error scanning the rows in GetAvailability() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		availabilities = append(availabilities, availability)
	}

	return c.JSON(availabilities)
}

// UpdateAvailability met à jour une disponibilité existante dans la base de données.
func UpdateAvailability(c *fiber.Ctx) error {
	id := c.Query("id")

	var availability Availability
	if err := c.BodyParser(&availability); err != nil {
		fmt.Println("💥 Error parsing the body in UpdateAvailability() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	stmt, err := db.Prepare("UPDATE availability SET PhoneNumber=$1, Availability=$2, Duration=$3, Repeat=$4 WHERE ID=$5")
	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in UpdateAvailability() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the SQL statement in UpdateAvailability() : ", err)
			return
		}
	}(stmt)

	_, err = stmt.Exec(availability.PhoneNumber, availability.Availability, availability.Duration, availability.Repeat, id)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in UpdateAvailability() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteAvailability supprime une disponibilité de la base de données.
func DeleteAvailability(c *fiber.Ctx) error {
	id := c.Query("id")

	stmt, err := db.Prepare("DELETE FROM availability WHERE ID=$1")
	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in DeleteAvailability() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the SQL statement in DeleteAvailability() : ", err)
			return
		}
	}(stmt)

	_, err = stmt.Exec(id)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in DeleteAvailability() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
