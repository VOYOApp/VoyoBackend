package main

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"math/rand"
	"strconv"
	"strings"
)

func CreateVisit(c *fiber.Ctx) error {
	if c.Locals("user").(*CustomClaims).Role == "VISITOR" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized access. You need to be a prospect to create a visit.",
		})
	}

	type VisitToCreate struct {
		Visit
		Criterias []Criteria `json:"criterias"`
	}

	var vtc VisitToCreate
	if err := c.BodyParser(&vtc); err != nil {
		fmt.Println("💥 Error parsing the body in CreateVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	// TODO: remove the price from this request and calculate it on the backend to prevent any manipulation from the user
	if vtc.PhoneNumberVisitor == "" || vtc.StartTime.IsZero() || vtc.Price == 0 || (vtc.IdAddressGMap == "" && (vtc.X == 0 && vtc.Y == 0)) || vtc.IdTypeRealEstate == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Please provide all the required fields.",
		})
	}

	// TODO: x or y is empty and if so fill them by doing a request to gmap

	// Check if the user exists
	if !checkUserExists(vtc.PhoneNumberVisitor) || !checkUserRole(vtc.PhoneNumberVisitor, "VISITOR") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "The visitor does not exist or is not a visitor.",
		})
	}

	// TODO: Check if the user is available or not

	// 1) Prepare the request
	stmt, err := db.Prepare("INSERT INTO visit (phonenumberprospect, phonenumbervisitor, codeverification, starttime, price, status, note, idaddressgmap, idtyperealestate, x, y) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)")
	if err != nil {
		fmt.Println("💥 Error preparing the SQL statement in CreateVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	// Generate a random codeverification number for the visit
	vtc.CodeVerification = rand.Intn(999999-100000) + 100000
	vtc.Status = "PENDING"
	// 2) Execute the request
	_, err = stmt.Exec(c.Locals("user").(*CustomClaims).PhoneNumber, vtc.PhoneNumberVisitor, vtc.CodeVerification, vtc.StartTime, vtc.Price, vtc.Status, vtc.Note, vtc.IdAddressGMap, vtc.IdTypeRealEstate, vtc.X, vtc.Y)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in CreateVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	// 3) Get the ID of the visit
	row := db.QueryRow("SELECT idvisit FROM visit WHERE phonenumbervisitor = $1 AND starttime = $2", vtc.PhoneNumberVisitor, vtc.StartTime)
	if err != nil {
		fmt.Println("💥 Error querying the database in CreateVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	var id int
	err = row.Scan(&id)
	if err != nil {
		fmt.Println("💥 Error scanning the row in CreateVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	// Loop through the criterias and insert them in the database
	for _, crit := range vtc.Criterias {
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
			crit.Criteria,
			crit.CriteriaAnswer,
			crit.PhotoRequired,
			crit.Photo,
			crit.VideoRequired,
			crit.Video,
			c.Locals("user").(*CustomClaims).PhoneNumber, // Retrieve the phone number from the context (middleware
			crit.Reusable,
		)
		if err != nil {
			fmt.Println("💥 Error executing the SQL statement in CreateCriteria() : ", err)
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "The visit already exists.",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		// Get the new criteria id
		row := db.QueryRow("SELECT idcriteria FROM criteria WHERE criteria = $1 AND criteriaanswer = $2 AND photo = $3 AND video = $4 AND phonenumber = $5", crit.Criteria, crit.CriteriaAnswer, crit.Photo, crit.Video, c.Locals("user").(*CustomClaims).PhoneNumber)
		if err != nil {
			fmt.Println("💥 Error querying the database in CreateVisit() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		var idCriteria int
		err = row.Scan(&idCriteria)
		if err != nil {
			fmt.Println("💥 Error scanning the row in CreateVisit() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		// Insert the link between the criteria and the visit
		stmt, err = db.Prepare("INSERT INTO linkcriteriavisit (idcriteria, idvisit) VALUES ($1, $2)")
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

		_, err = stmt.Exec(idCriteria, id)
		if err != nil {
			fmt.Println("💥 Error executing the SQL statement in CreateVisit() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
	}

	return c.Status(fiber.StatusCreated).SendString("Visit created successfully")
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
	searchString2 := "phoneNumber"
	if role == "PROSPECT" {
		searchString = "PhoneNumberProspect"
		searchString2 = "PhoneNumberVisitor"
	} else {
		searchString = "PhoneNumberVisitor"
		searchString2 = "PhoneNumberProspect"
	}

	isNot := ""

	if visitType == "PASSED" {
		isNot = "NOT"
	}

	request := fmt.Sprintf(`
		SELECT FirstName, UPPER(CONCAT(LEFT(LastName, 1), '.')) AS LastName, visit.idaddressgmap, StartTime, (starttime + duration) AS EndTime, Duration, visit.status, Note, visit.idvisit
		FROM visit
		         JOIN public."user" u ON visit.%s = u.phonenumber
		         JOIN typerealestate t ON visit.idtyperealestate = t.idtyperealestate
		WHERE %s = '%s' AND visit.Status %s IN ('PENDING', 'ACCEPTED')
		ORDER BY StartTime ASC
`, searchString2, searchString, phoneNumber, isNot)

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
		EndTime       string  `json:"endTime"`
		Duration      string  `json:"duration"`
		Status        string  `json:"status"`
		Note          float64 `json:"note"`
		IDVisit       int     `json:"idVisit"`
	}

	var visits []upcomingVisits

	for rows.Next() {
		var visit upcomingVisits
		err := rows.Scan(&visit.FirstName, &visit.LastName, &visit.IdAddressGmap, &visit.StartTime, &visit.EndTime, &visit.Duration, &visit.Status, &visit.Note, &visit.IDVisit)
		if err != nil {
			fmt.Println("💥 Error scanning the rows in GetUpcomingVisits() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		//TODO Améliorer
		gmap, _ := getAddressFromGMapsID(visit.IdAddressGmap)
		if len(gmap.Results) > 0 {
			visit.Address = gmap.Results[0].FormattedAddress
		}

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
			       v.idaddressgmap,
			       Date(StartTime)                                                       AS Date,
			       TO_CHAR(StartTime, 'HH24hMI')                                         AS StartTime,
			       TO_CHAR(starttime + tr.duration, 'HH24hMI')                           AS EndTime,
			       tr.duration,
			       v.Status,
			       FirstName,
			       UPPER(CONCAT(LEFT(LastName, 1), '.'))                                 AS LastName,
			       profilepicture,
			       vc.count                                                              AS VisitCount,
			       COALESCE(navg.avg, 0)                                                              AS NoteAvg,
			       price,
			       note,
			       v.codeverification,
			       CASE WHEN v.status NOT IN ('DONE', 'ACCEPTED') THEN FALSE ELSE TRUE END AS VisitAccepted,
			       CASE
			           WHEN (SELECT COUNT(idVisit) FROM public.linkcriteriavisit WHERE idVisit = %[1]s) > 0 THEN TRUE
			           ELSE FALSE END                                                    AS CriteriaSent
			FROM visit v
			         JOIN public.typerealestate tr ON v.idtyperealestate = tr.idtyperealestate
			         JOIN public."user" u ON v.phonenumbervisitor = u.phonenumber
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
			err := row.Scan(&visit.Visit.IDVisit, &visit.Visit.Address.IdAddressGmap, &visit.Visit.Details.Date, &visit.Visit.Details.StartTime, &visit.Visit.Details.EndTime, &visit.Visit.Details.Duration, &visit.Visit.Details.Status, &visit.Visitor.FirstName, &visit.Visitor.LastName, &visit.Visitor.ProfilePicture, &visit.Visitor.VisitCount, &visit.Visitor.NoteAVG, &visit.Visit.Details.Price, &visit.Visit.Details.Note, &visit.Visit.Details.Code, &visit.Visit.Details.VisitAccepted, &visit.Visit.Details.CriteriaSent)
			if err != nil {
				fmt.Println("💥 Error scanning the row in GetVisit() : ", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "An error has occurred, please try again later.",
				})
			}

			visit.Visit.Address.googleMapsResponse, _ = getAddressFromGMapsID(visit.Visit.Address.IdAddressGmap)

			if c.Locals("user").(*CustomClaims).Role == "VISITOR" {
				visit.Visit.Details.Code = 0
			}

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
		// TODO: adapt the values retrieved to the new db
		if c.Locals("user").(*CustomClaims).Role == "ADMIN" {
			rows, err := db.Query("SELECT idvisit, phonenumberprospect, phonenumbervisitor, codeverification, starttime, price, status, note FROM visit")
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
				err := rows.Scan(&visit.IdVisit, &visit.PhoneNumberProspect, &visit.PhoneNumberVisitor, &visit.CodeVerification, &visit.StartTime, &visit.Price, &visit.Status, &visit.Note)
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

func GetVisitVerificationCode(c *fiber.Ctx) error {
	idVisit := c.Query("idVisit")

	if idVisit == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Please provide the ID of the visit",
		})
	}

	if !hasAuthorizedVisitAccess(c.Locals("user").(*CustomClaims).PhoneNumber, idVisit) || c.Locals("user").(*CustomClaims).Role == "VISITOR" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized access",
		})
	}

	row := db.QueryRow("SELECT codeverification FROM visit WHERE idvisit = $1", idVisit)

	var code int
	err := row.Scan(&code)
	if err != nil {
		fmt.Println("💥 Error scanning the row in GetVisitVerificationCode() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.JSON(fiber.Map{
		"code": code,
	})
}

func CheckVisitVerificationCode(c *fiber.Ctx) error {
	idVisit := c.Query("idVisit")
	code := c.Query("code")

	if idVisit == "" || code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Please provide the ID of the visit and the verification code",
		})
	}

	if !hasAuthorizedVisitAccess(c.Locals("user").(*CustomClaims).PhoneNumber, idVisit) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized access",
		})
	}

	row := db.QueryRow("SELECT codeverification FROM visit WHERE idvisit = $1", idVisit)

	var dbCode int
	err := row.Scan(&dbCode)
	if err != nil {
		fmt.Println("💥 Error scanning the row in CheckVisitVerificationCode() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	if strconv.Itoa(dbCode) == code {
		return c.SendStatus(fiber.StatusNoContent)
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": "The verification code is incorrect",
	})
}
