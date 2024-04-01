package main

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
)

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
		c.Locals("user").(*CustomClaims).PhoneNumber, // Retrieve the phone number from the context (middleware
		criteria.Reusable,
	)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in CreateCriteria() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusCreated)
}

func GetCriteria(c *fiber.Ctx) error {
	id := c.Query("id")
	idVisit := c.Query("idVisit")

	// Prepare SQL statement
	var query string
	var args []interface{}
	if id != "" && hasAuthorizedCriteriaAccess(c.Locals("user").(*CustomClaims).PhoneNumber, id) {
		query = `
            SELECT idCriteria, criteria, criteriaAnswer, photoRequired, photo, videoRequired, video, reusable
            FROM criteria
            WHERE idCriteria = $1
        `
		args = []interface{}{id}
	} else if idVisit != "" && hasAuthorizedVisitAccess(c.Locals("user").(*CustomClaims).PhoneNumber, idVisit) {
		query = `
            SELECT idCriteria, criteria, criteriaAnswer, photoRequired, photo, videoRequired, video, reusable
            FROM criteria
            WHERE idCriteria IN (
                SELECT idCriteria
                FROM linkcriteriavisit
                WHERE idVisit=$1
            )
        `
		args = []interface{}{idVisit}
	} else {
		query = `
            SELECT idCriteria, criteria, criteriaAnswer, photoRequired, photo, videoRequired, video, reusable
            FROM criteria
            WHERE phoneNumber=$1 AND reusable=true
        `
		args = []interface{}{c.Locals("user").(*CustomClaims).PhoneNumber}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Println("Error querying the database in GetCriteria():", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println("Error closing the rows in GetCriteria():", err)
			return
		}
	}(rows)

	var criterias []Criteria
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
			&criteria.Reusable,
		)
		if err != nil {
			log.Println("Error scanning the row in GetCriteria():", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		criterias = append(criterias, criteria)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows in GetCriteria():", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.JSON(criterias)
}

func UpdateCriteria(c *fiber.Ctx) error {
	id := c.Query("id")

	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Please provide the ID of the visit to update",
		})
	}

	if !hasAuthorizedCriteriaAccess(c.Locals("user").(*CustomClaims).PhoneNumber, id) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized access",
		})
	}

	var criteria Criteria
	if err := c.BodyParser(&criteria); err != nil {
		fmt.Println("💥 Error parsing the body in UpdateCriteria() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	var updateQuery string
	var args []interface{}

	placeholderIndex := 1 // Start with placeholder index 1

	if criteria.Criteria != "" {
		updateQuery += fmt.Sprintf("criteria=$%d, ", placeholderIndex)
		args = append(args, criteria.Criteria)
		placeholderIndex++
	}

	if criteria.CriteriaAnswer != "" {
		updateQuery += fmt.Sprintf("criteriaAnswer=$%d, ", placeholderIndex)
		args = append(args, criteria.CriteriaAnswer)
		placeholderIndex++
	}

	// TODO: Investigate because it may not work if the photoRequired is false
	if criteria.PhotoRequired {
		updateQuery += fmt.Sprintf("photoRequired=$%d, ", placeholderIndex)
		args = append(args, criteria.PhotoRequired)
		placeholderIndex++
	}

	if criteria.Photo != "" {
		updateQuery += fmt.Sprintf("photo=$%d, ", placeholderIndex)
		args = append(args, criteria.Photo)
		placeholderIndex++
	}

	// TODO: Investigate because it may not work if the videoRequired is false
	if criteria.VideoRequired {
		updateQuery += fmt.Sprintf("videoRequired=$%d, ", placeholderIndex)
		args = append(args, criteria.VideoRequired)
		placeholderIndex++
	}

	if criteria.Video != "" {
		updateQuery += fmt.Sprintf("video=$%d, ", placeholderIndex)
		args = append(args, criteria.Video)
		placeholderIndex++
	}

	if criteria.Reusable {
		updateQuery += fmt.Sprintf("reusable=$%d, ", placeholderIndex)
		args = append(args, criteria.Reusable)
		placeholderIndex++
	}

	if criteria.PhoneNumber != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "You are not allowed to update the phone number",
		})
	}

	// Remove the trailing comma and space
	updateQuery = updateQuery[:len(updateQuery)-2]

	stmt, err := db.Prepare(fmt.Sprintf(`
		UPDATE criteria
		SET %s
		WHERE idCriteria=$%d
	`, updateQuery, placeholderIndex))

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

	args = append(args, id)
	_, err = stmt.Exec(args...)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in UpdateCriteria() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
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
