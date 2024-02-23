package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func CreateLinkCriteriaVisit(c *fiber.Ctx) error {
	var linkCriteriaVisit LinkCriteriaVisit
	if err := c.BodyParser(&linkCriteriaVisit); err != nil {
		fmt.Println("💥 Error parsing the body in CreateLinkCriteriaVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	stmt, err := db.Prepare(`
		INSERT INTO linkCriteriaVisit (idCriteria, idVisit)
		VALUES ($1, $2)
	`)
	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in CreateLinkCriteriaVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in CreateLinkCriteriaVisit() : ", err)
			return
		}
	}(stmt)

	_, err = stmt.Exec(linkCriteriaVisit.IDCriteria, linkCriteriaVisit.IDVisit)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in CreateLinkCriteriaVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.Status(fiber.StatusCreated).SendString("LinkCriteriaVisit created successfully")
}

func GetLinkCriteriaVisit(c *fiber.Ctx) error {
	idCriteria := c.Query("idCriteria")
	idVisit := c.Query("idVisit")

	// If both ID criteria and ID visit are specified in the query parameters,
	// retrieve only that specific linkCriteriaVisit.
	if idCriteria != "" && idVisit != "" {
		var linkCriteriaVisit LinkCriteriaVisit
		stmt, err := db.Prepare(`
			SELECT idCriteria, idVisit
			FROM linkCriteriaVisit
			WHERE idCriteria = $1 AND idVisit = $2
		`)
		if err != nil {
			fmt.Println("💥 Error preparing the SQL statement in GetLinkCriteriaVisit() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		defer func(stmt *sql.Stmt) {
			err := stmt.Close()
			if err != nil {
				fmt.Println("💥 Error closing the statement in GetLinkCriteriaVisit() : ", err)
				return
			}
		}(stmt)

		row := stmt.QueryRow(idCriteria, idVisit)
		err = row.Scan(&linkCriteriaVisit.IDCriteria, &linkCriteriaVisit.IDVisit)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "LinkCriteriaVisit not found",
				})
			}

			fmt.Println("💥 Error scanning the row in GetLinkCriteriaVisit() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		return c.JSON(linkCriteriaVisit)
	}

	// If only ID criteria is specified, return an error
	if idCriteria != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Both idCriteria and idVisit must be specified or omitted together.",
		})
	}

	// If only ID visit is specified, retrieve all idCriteria for that idVisit.
	if idVisit != "" {
		rows, err := db.Query(`
			SELECT idCriteria, idVisit
			FROM linkCriteriaVisit
			WHERE idVisit = $1
		`, idVisit)
		if err != nil {
			fmt.Println("💥 Error querying the database in GetLinkCriteriaVisit() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				fmt.Println("💥 Error closing the rows in GetLinkCriteriaVisit() : ", err)
				return
			}
		}(rows)

		var linkCriteriaVisitList []LinkCriteriaVisit
		for rows.Next() {
			var linkCriteriaVisit LinkCriteriaVisit
			err := rows.Scan(&linkCriteriaVisit.IDCriteria, &linkCriteriaVisit.IDVisit)
			if err != nil {
				fmt.Println("💥 Error scanning the rows in GetLinkCriteriaVisit() : ", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "An error has occurred, please try again later.",
				})
			}
			linkCriteriaVisitList = append(linkCriteriaVisitList, linkCriteriaVisit)
		}

		return c.JSON(linkCriteriaVisitList)
	}

	// If no ID criteria or ID visit is specified, retrieve all linkCriteriaVisit.
	rows, err := db.Query(`
		SELECT idCriteria, idVisit
		FROM linkCriteriaVisit
	`)
	if err != nil {
		fmt.Println("💥 Error querying the database in GetLinkCriteriaVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println("💥 Error closing the rows in GetLinkCriteriaVisit() : ", err)
			return
		}
	}(rows)

	var linkCriteriaVisitList []LinkCriteriaVisit
	for rows.Next() {
		var linkCriteriaVisit LinkCriteriaVisit
		err := rows.Scan(&linkCriteriaVisit.IDCriteria, &linkCriteriaVisit.IDVisit)
		if err != nil {
			fmt.Println("💥 Error scanning the rows in GetLinkCriteriaVisit() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		linkCriteriaVisitList = append(linkCriteriaVisitList, linkCriteriaVisit)
	}

	return c.JSON(linkCriteriaVisitList)
}

// DeleteLinkCriteriaVisit deletes a linkCriteriaVisit from the database
func DeleteLinkCriteriaVisit(c *fiber.Ctx) error {
	idCriteria := c.Query("idCriteria")
	idVisit := c.Query("idVisit")

	stmt, err := db.Prepare(`
		DELETE FROM linkCriteriaVisit
		WHERE idCriteria = $1 AND idVisit = $2
	`)
	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in DeleteLinkCriteriaVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in DeleteLinkCriteriaVisit() : ", err)
			return
		}
	}(stmt)

	_, err = stmt.Exec(idCriteria, idVisit)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in DeleteLinkCriteriaVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
