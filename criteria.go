package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

// CreateCriteria creates a new criteria entry in the database
func CreateCriteria(c *fiber.Ctx) error {
	var criteria Criteria
	if err := c.BodyParser(&criteria); err != nil {
		fmt.Println("💥 Error parsing the body in CreateCriteria() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	stmt, err := db.Prepare(`
		INSERT INTO criteria (criteria, criteriaAnswer, photoRequired, photo, videoRequired, video, phoneNumber, reusable)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`)

	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in CreateCriteria() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in CreateCriteria() : ", err)
			return
		}
	}(stmt)

	_, err = stmt.Exec(
		criteria.Criteria,
		criteria.CriteriaAnswer,
		criteria.PhotoRequired,
		criteria.Photo,
		criteria.VideoRequired,
		criteria.Video,
		criteria.PhoneNumber,
		criteria.Reusable,
	)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in CreateCriteria() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.Status(fiber.StatusCreated).SendString("Criteria created successfully")
}

// GetCriteria retrieves criteria based on ID or all criteria if no ID is specified
func GetCriteria(c *fiber.Ctx) error {
	id := c.Query("id")

	// If an ID is specified in the query parameters,
	// retrieve only that specific criteria.
	if id != "" {
		var criteria Criteria
		stmt, err := db.Prepare(`
			SELECT idCriteria, criteria, criteriaAnswer, photoRequired, photo, videoRequired, video, phoneNumber, reusable
			FROM criteria
			WHERE idCriteria = $1
		`)
		if err != nil {
			fmt.Println("💥 Error preparing the SQL statement in GetCriteria() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		defer func(stmt *sql.Stmt) {
			err := stmt.Close()
			if err != nil {
				fmt.Println("💥 Error closing the statement in GetCriteria() : ", err)
				return
			}
		}(stmt)

		row := stmt.QueryRow(id)
		err = row.Scan(
			&criteria.ID,
			&criteria.Criteria,
			&criteria.CriteriaAnswer,
			&criteria.PhotoRequired,
			&criteria.Photo,
			&criteria.VideoRequired,
			&criteria.Video,
			&criteria.PhoneNumber,
			&criteria.Reusable,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Criteria not found",
				})
			}

			fmt.Println("💥 Error scanning the row in GetCriteria() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		return c.JSON(criteria)
	}

	// If no ID is specified, retrieve all criteria.
	rows, err := db.Query(`
		SELECT idCriteria, criteria, criteriaAnswer, photoRequired, photo, videoRequired, video, phoneNumber, reusable
		FROM criteria
	`)
	if err != nil {
		fmt.Println("💥 Error querying the database in GetCriteria() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println("💥 Error closing the rows in GetCriteria() : ", err)
			return
		}
	}(rows)

	var criteriaList []Criteria
	for rows.Next() {
		var criteria Criteria
		err := rows.Scan(
			&criteria.ID,
			&criteria.Criteria,
			&criteria.CriteriaAnswer,
			&criteria.PhotoRequired,
			&criteria.Photo,
			&criteria.VideoRequired,
			&criteria.Video,
			&criteria.PhoneNumber,
			&criteria.Reusable,
		)
		if err != nil {
			fmt.Println("💥 Error scanning the rows in GetCriteria() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		criteriaList = append(criteriaList, criteria)
	}

	return c.JSON(criteriaList)
}

// UpdateCriteria updates an existing criteria in the database
func UpdateCriteria(c *fiber.Ctx) error {
	id := c.Query("id")

	var criteria Criteria
	if err := c.BodyParser(&criteria); err != nil {
		fmt.Println("💥 Error parsing the body in UpdateCriteria() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	stmt, err := db.Prepare(`
		UPDATE criteria
		SET criteria=$1, criteriaAnswer=$2, photoRequired=$3, photo=$4, videoRequired=$5, video=$6, phoneNumber=$7, reusable=$8
		WHERE idCriteria=$9
	`)
	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in UpdateCriteria() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in UpdateCriteria() : ", err)
			return
		}
	}(stmt)

	_, err = stmt.Exec(
		criteria.Criteria,
		criteria.CriteriaAnswer,
		criteria.PhotoRequired,
		criteria.Photo,
		criteria.VideoRequired,
		criteria.Video,
		criteria.PhoneNumber,
		criteria.Reusable,
		id,
	)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in UpdateCriteria() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteCriteria deletes a criteria from the database
func DeleteCriteria(c *fiber.Ctx) error {
	id := c.Query("id")

	stmt, err := db.Prepare(`
		DELETE FROM criteria
		WHERE idCriteria=$1
	`)
	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in DeleteCriteria() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in DeleteCriteria() : ", err)
			return
		}
	}(stmt)

	_, err = stmt.Exec(id)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in DeleteCriteria() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
