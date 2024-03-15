package main

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"strings"
)

// CreateVisit crée une nouvelle visite dans la base de données.
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

// GetVisit récupère une visite spécifique à partir de son ID, ou toutes les visites s'il n'y a pas d'ID spécifié.
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
			err := row.Scan(&visit.Visit.IDVisit, &visit.Visit.Address.IdAddressGmap, &visit.Visit.Details.Date, &visit.Visit.Details.StartTime, &visit.Visit.Details.EndTime, &visit.Visit.Details.Duration, &visit.Visit.Details.Status, &visit.Visitor.FirstName, &visit.Visitor.LastName, &visit.Visitor.ProfilePicture, &visit.Visitor.VisitCount, &visit.Visitor.NoteAVG, &visit.Visit.Details.Price, &visit.Visit.Details.VisitAccepted, &visit.Visit.Details.CriteriaSent)
			if err != nil {
				fmt.Println("💥 Error scanning the row in GetVisit() : ", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "An error has occurred, please try again later.",
				})
			}

			visit.Visit.Address.Address, _ = getAddressFromGMapsID(visit.Visit.Address.IdAddressGmap)

			return c.JSON(visit)
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized access",
			})
		}
	} else {
		// Return all the visits
		if c.Locals("user").(*CustomClaims).Role == "ADMIN" {
			rows, err := db.Query("SELECT * FROM visit")
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

// UpdateVisit met à jour une visite existante dans la base de données.
func UpdateVisit(c *fiber.Ctx) error {
	id := c.Query("id")

	var visit Visit
	if err := c.BodyParser(&visit); err != nil {
		fmt.Println("💥 Error parsing the body in UpdateVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	stmt, err := db.Prepare("UPDATE visit SET PhoneNumberProspect=$1, PhoneNumberVisitor=$2, IdRealEstate=$3, CodeVerification=$4, StartTime=$5, Price=$6, Status=$7, Note=$8 WHERE ID=$9")
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
			return
		}
	}(stmt)

	_, err = stmt.Exec(visit.PhoneNumberProspect, visit.PhoneNumberVisitor, visit.IdRealEstate, visit.CodeVerification, visit.StartTime, visit.Price, visit.Status, visit.Note, id)
	if err != nil {
		fmt.Println("💥 Error executing the SQL statement in UpdateVisit() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusOK)
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

	if visitType != "UPCOMING" {
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

		visit.Address, _ = getAddressFromGMapsID(visit.IdAddressGmap)

		visits = append(visits, visit)
	}

	return c.JSON(visits)
}
