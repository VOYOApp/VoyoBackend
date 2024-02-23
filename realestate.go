package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

// CreateRealEstate crée un nouveau bien immobilier dans la base de données.
func CreateRealEstate(c *fiber.Ctx) error {
	var realEstate RealEstate
	if err := c.BodyParser(&realEstate); err != nil {
		fmt.Println("💥 Error parsing the body in CreateRealEstate() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	stmt, err := db.Prepare("INSERT INTO realEstate (IdAddressGMap, IdTypeRealEstate) VALUES ($1, $2)")
	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in CreateRealEstate() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in CreateRealEstate() : ", err)
			return
		}
	}(stmt)

	_, err = stmt.Exec(realEstate.IdAddressGMap, realEstate.IdTypeRealEstate)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in CreateRealEstate() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.Status(fiber.StatusCreated).SendString("Bien immobilier créé avec succès")
}

// GetRealEstate récupère un bien immobilier spécifique à partir de son ID, ou tous les biens immobiliers s'il n'y a pas d'ID spécifié.
func GetRealEstate(c *fiber.Ctx) error {
	id := c.Query("id")

	// Si un ID est spécifié dans les paramètres de la requête,
	// on récupère uniquement ce bien immobilier spécifique.
	if id != "" {
		var realEstate RealEstate
		stmt, err := db.Prepare("SELECT IdRealEstate, IdAddressGMap, IdTypeRealEstate FROM realEstate WHERE IdRealEstate = $1")
		if err != nil {
			fmt.Println("💥 Error preparing the SQL statement in GetRealEstate() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		defer func(stmt *sql.Stmt) {
			err := stmt.Close()
			if err != nil {
				fmt.Println("💥 Error closing the statement in GetRealEstate() : ", err)
				return
			}
		}(stmt)

		row := stmt.QueryRow(id)
		err = row.Scan(&realEstate.IdRealEstate, &realEstate.IdAddressGMap, &realEstate.IdTypeRealEstate)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Real estate not found",
				})
			}

			fmt.Println("💥 Error scanning the row in GetRealEstate() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		return c.JSON(realEstate)
	}

	// Si aucun ID n'est spécifié, on récupère tous les biens immobiliers.
	rows, err := db.Query("SELECT IdRealEstate, IdAddressGMap, IdTypeRealEstate FROM realEstate")
	if err != nil {
		fmt.Println("💥 Error querying the database in GetRealEstate() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println("💥 Error closing the rows in GetRealEstate() : ", err)
			return
		}
	}(rows)

	var realEstates []RealEstate
	for rows.Next() {
		var realEstate RealEstate
		err := rows.Scan(&realEstate.IdRealEstate, &realEstate.IdAddressGMap, &realEstate.IdTypeRealEstate)
		if err != nil {
			fmt.Println("💥 Error scanning the rows in GetRealEstate() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		realEstates = append(realEstates, realEstate)
	}

	return c.JSON(realEstates)
}

// UpdateRealEstate met à jour un bien immobilier existant dans la base de données.
func UpdateRealEstate(c *fiber.Ctx) error {
	id := c.Query("id")

	var realEstate RealEstate
	if err := c.BodyParser(&realEstate); err != nil {
		fmt.Println("💥 Error parsing the body in UpdateRealEstate() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	stmt, err := db.Prepare("UPDATE realEstate SET IdAddressGMap=$1, IdTypeRealEstate=$2 WHERE IdRealEstate=$3")
	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in UpdateRealEstate() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in UpdateRealEstate() : ", err)
			return
		}
	}(stmt)

	_, err = stmt.Exec(realEstate.IdAddressGMap, realEstate.IdTypeRealEstate, id)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in UpdateRealEstate() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteRealEstate supprime un bien immobilier de la base de données.
func DeleteRealEstate(c *fiber.Ctx) error {
	id := c.Query("id")

	stmt, err := db.Prepare("DELETE FROM realEstate WHERE IdRealEstate=$1")
	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in DeleteRealEstate() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in DeleteRealEstate() : ", err)
			return
		}
	}(stmt)

	_, err = stmt.Exec(id)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in DeleteRealEstate() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
