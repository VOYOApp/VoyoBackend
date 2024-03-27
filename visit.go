package main

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"strings"
)

func CreateVisit(c *fiber.Ctx) error {
	var visit Visit
	if err := c.BodyParser(&visit); err != nil {
		fmt.Println("💥 Error parsing the body in CreateVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	stmt, err := db.Prepare("INSERT INTO visit (PhoneNumberProspect, PhoneNumberVisitor, IdRealEstate, CodeVerification, StartTime, Price, Status, Note) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")
	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in CreateVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in CreateVisit() : ", err)
			return
		}
	}(stmt)

	_, err = stmt.Exec(visit.PhoneNumberProspect, visit.PhoneNumberVisitor, visit.IdRealEstate, visit.CodeVerification, visit.StartTime, visit.Price, visit.Status, visit.Note)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in CreateVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.Status(fiber.StatusCreated).SendString("Visite créée avec succès")
}

// DeleteVisit supprime une visite de la base de données.
func DeleteVisit(c *fiber.Ctx) error {
	id := c.Query("id")

	stmt, err := db.Prepare("DELETE FROM visit WHERE ID=$1")
	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in DeleteVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in DeleteVisit() : ", err)
			return
		}
	}(stmt)

	_, err = stmt.Exec(id)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in DeleteVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ============================================= THIS IS CLEAN BELOW ============================================= //

func GetVisitsList(c *fiber.Ctx) error {
	// Get the number of users
	phoneNumber := c.Locals("user").(*CustomClaims).PhoneNumber
	role := c.Locals("user").(*CustomClaims).Role

	visitType := strings.ToUpper(c.Query("type"))

	searchString := "phoneNumber"
	if role == "PROSPECT" {
		searchString = "PhoneNumberProspect"
	} else {
		searchString = "PhoneNumberVisitor"
	}

	isNot := ""

	if visitType != "PASSED" {
		isNot = "NOT"
	}

	request := fmt.Sprintf(`
		SELECT FirstName, UPPER(CONCAT(LEFT(LastName, 1), '.')) AS LastName, r.IdAddressGmap, StartTime, Status, Note, visit.idvisit
		FROM visit
		         JOIN public."user" u ON visit.phonenumberprospect = u.phonenumber
		         JOIN public.realestate r ON r.idrealestate = visit.idrealestate
		WHERE %s = '%s' AND Status %s IN ('PENDING', 'ACCEPTED')
`, searchString, phoneNumber, isNot)

	rows, err := db.Query(request)
	if err != nil {
		fmt.Println("💥 Error querying the database in GetUpcomingVisits() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println("💥 Error closing the rows in GetUpcomingVisits() : ", err)
			return
		}
	}(rows)

	type upcomingVisits struct {
		FirstName     string  `json:"firstName"`
		LastName      string  `json:"lastName"`
		IdAddressGmap string  `json:"idAddressGmap"`
		Address       string  `json:"address"`
		StartTime     string  `json:"startTime"`
		Status        string  `json:"status"`
		Note          float64 `json:"note"`
		IDVisit       int     `json:"idVisit"`
	}

	var visits []upcomingVisits

	for rows.Next() {
		var visit upcomingVisits
		err := rows.Scan(&visit.FirstName, &visit.LastName, &visit.IdAddressGmap, &visit.StartTime, &visit.Status, &visit.Note, &visit.IDVisit)
		if err != nil {
			fmt.Println("💥 Error scanning the rows in GetUpcomingVisits() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		gmap, _ := getAddressFromGMapsID(visit.IdAddressGmap)
		visit.Address = gmap.Results[0].FormattedAddress

		visits = append(visits, visit)
	}

	return c.JSON(visits)
}

func GetVisit(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Query("id"))

	if id != "" {
		if hasAuthorizedVisitAccess(c.Locals("user").(*CustomClaims).PhoneNumber, id) {
			request := fmt.Sprintf(`
			SELECT idvisit,
			       r.IdAddressGmap,
			       Date(StartTime)                                                       AS Date,
			       TO_CHAR(StartTime, 'HH24hMI')                                         AS StartTime,
			       TO_CHAR(starttime + tr.duration, 'HH24hMI')                           AS EndTime,
			       tr.duration,
			       Status,
			       FirstName,
			       UPPER(CONCAT(LEFT(LastName, 1), '.'))                                 AS LastName,
			       profilepicture,
			       vc.count                                                              AS VisitCount,
			       navg.avg                                                              AS NoteAvg,
			       price,
			       note,
			       CASE WHEN status NOT IN ('DONE', 'ACCEPTED') THEN FALSE ELSE TRUE END AS VisitAccepted,
			       CASE
			           WHEN (SELECT COUNT(idVisit) FROM public.linkcriteriavisit WHERE idVisit = 135) > 0 THEN TRUE
			           ELSE FALSE END                                                    AS CriteriaSent
			FROM visit
			         JOIN public.realestate r ON r.idrealestate = visit.idrealestate
			         JOIN public.typerealestate tr ON tr.idtyperealestate = r.idtyperealestate
			         JOIN public."user" u ON visit.phonenumbervisitor = u.phonenumber
			         JOIN (SELECT COUNT(idvisit) AS count
			               FROM public.visit
			               WHERE phonenumbervisitor = (SELECT phonenumbervisitor FROM visit WHERE idvisit = %[1]s)
			                 AND status = 'DONE') AS vc ON TRUE
			         JOIN (SELECT AVG(note) AS avg
			               FROM public.visit
			               WHERE phonenumbervisitor = (SELECT phonenumbervisitor FROM visit WHERE idvisit = %[1]s)
			                 AND status = 'DONE'
			                 AND note != 0.0) AS navg ON TRUE
			WHERE idvisit = %[1]s;`,
				id)

			row := db.QueryRow(request)

			var visit visitDetails
			err := row.Scan(&visit.Visit.IDVisit, &visit.Visit.Address.IdAddressGmap, &visit.Visit.Details.Date, &visit.Visit.Details.StartTime, &visit.Visit.Details.EndTime, &visit.Visit.Details.Duration, &visit.Visit.Details.Status, &visit.Visitor.FirstName, &visit.Visitor.LastName, &visit.Visitor.ProfilePicture, &visit.Visitor.VisitCount, &visit.Visitor.NoteAVG, &visit.Visit.Details.Price, &visit.Visit.Details.Note, &visit.Visit.Details.VisitAccepted, &visit.Visit.Details.CriteriaSent)
			if err != nil {
				fmt.Println("💥 Error scanning the row in GetVisit() : ", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "An error has occurred, please try again later.",
				})
			}

			visit.Visit.Address.googleMapsResponse, _ = getAddressFromGMapsID(visit.Visit.Address.IdAddressGmap)

			// Select all criterias for the visit
			rows, err := db.Query("SELECT criteria.idcriteria, criteria.criteria, criteriaanswer, photo, video, photorequired, videorequired FROM public.criteria join public.linkcriteriavisit on criteria.idcriteria = linkcriteriavisit.idcriteria where idvisit = $1", id)
			if err != nil {
				fmt.Println("💥 Error querying the database in GetVisit() : ", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "An error has occurred, please try again later.",
				})
			}

			defer func(rows *sql.Rows) {
				err := rows.Close()
				if err != nil {
					fmt.Println("💥 Error closing the rows in GetVisit() : ", err)
					return
				}
			}(rows)

			for rows.Next() {
				var crit Criteria
				err := rows.Scan(&crit.ID, &crit.Criteria, &crit.CriteriaAnswer, &crit.Photo, &crit.Video, &crit.PhotoRequired, &crit.VideoRequired)
				if err != nil {
					fmt.Println("💥 Error scanning the rows in GetVisit() : ", err)
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "An error has occurred, please try again later.",
					})
				}

				visit.Visit.Criterias = append(visit.Visit.Criterias, crit)
			}

			return c.JSON(visit)
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized access",
			})
		}
	} else {
		// Return all the visits
		if c.Locals("user").(*CustomClaims).Role == "ADMIN" {
			rows, err := db.Query("SELECT idvisit, phonenumberprospect, phonenumbervisitor, idrealestate, codeverification, starttime, price, status, note FROM visit")
			if err != nil {
				fmt.Println("💥 Error querying the database in GetVisit() : ", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "An error has occurred, please try again later.",
				})
			}

			defer func(rows *sql.Rows) {
				err := rows.Close()
				if err != nil {
					fmt.Println("💥 Error closing the rows in GetVisit() : ", err)
					return
				}
			}(rows)

			var visits []Visit
			for rows.Next() {
				var visit Visit
				err := rows.Scan(&visit.IdVisit, &visit.PhoneNumberProspect, &visit.PhoneNumberVisitor, &visit.IdRealEstate, &visit.CodeVerification, &visit.StartTime, &visit.Price, &visit.Status, &visit.Note)
				if err != nil {
					fmt.Println("💥 Error scanning the rows in GetVisit() : ", err)
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "An error has occurred, please try again later.",
					})
				}

				visits = append(visits, visit)
			}

			return c.JSON(visits)
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized access",
			})
		}
	}
}

func UpdateVisit(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Query("id"))

	/**
	* To ease the API creation process and protect the safety of the data, a visit can only be updated on 2
	* fields:
	* 	- Status
	* 	- Note
	*
	* The other fields are not supposed to be updated by the user. If needed, it will be implemented later.
	**/

	// TODO: to tightened the security even more, only the prospect should be able to change the note

	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Please provide the ID of the visit to update",
		})
	}

	if !hasAuthorizedVisitAccess(c.Locals("user").(*CustomClaims).PhoneNumber, id) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized access",
		})
	}

	var visit Visit
	if err := c.BodyParser(&visit); err != nil {
		fmt.Println("💥 Error parsing the body in UpdateVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	var updateQuery string
	var args []interface{}

	placeholderIndex := 1 // Start with placeholder index 1

	if visit.Status != "" {
		updateQuery += fmt.Sprintf("Status=$%d, ", placeholderIndex)
		args = append(args, visit.Status)
		placeholderIndex++
	}

	if visit.Note != 0 {
		updateQuery += fmt.Sprintf("Note=$%d, ", placeholderIndex)
		args = append(args, strconv.FormatFloat(visit.Note, 'f', -1, 64))
		placeholderIndex++
	}

	// Remove the trailing comma and space
	updateQuery = strings.TrimSuffix(updateQuery, ", ")

	query := fmt.Sprintf("UPDATE visit SET %s WHERE idvisit=$%d", updateQuery, len(args)+1)
	args = append(args, id)

	fmt.Println(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in UpdateVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in UpdateVisit() : ", err)
		}
	}(stmt)

	_, err = stmt.Exec(args...)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in UpdateVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
