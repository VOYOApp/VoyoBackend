package main

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"strconv"
	"strings"
)

// CreateUser crée un nouvel utilisateur dans la base de données.
func CreateUser(c *fiber.Ctx) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
		fmt.Println("💥 Error parsing the body in CreateUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("💥 Error hashing the password in CreateUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	stmt, err := db.Prepare(`INSERT INTO "user" (PhoneNumber, FirstName, LastName, Email, Password, IdRole, Biography, ProfilePicture, Pricing, IdAddressGMap, Radius, Status, CniFront, CniBack) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`)
	if err != nil {
		fmt.Println("💥 Error preparing the request in CreateUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in CreateUser()")
			return
		}
	}(stmt)

	status := "VALIDATED"

	if user.IdRole == 1 {
		status = "PENDING_VALIDATION"
	}

	_, err = stmt.Exec(user.PhoneNumber, user.FirstName, user.LastName, user.Email, hashedPassword, user.IdRole, user.Biography, user.ProfilePicture, user.Pricing, user.IdAddressGMap, user.Radius, status, user.CniFront, user.CniBack)
	if err != nil {
		fmt.Println("💥 Error executing the request in CreateUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	// Get the X and Y coordinates from the address using the google maps api
	if user.IdAddressGMap != nil {
		// Dereference the pointer and get coordinates
		address := utils.Trim(*user.IdAddressGMap, ' ')
		coordinates, errGmaps := getCoordinatesFromAddress(address)
		if errGmaps != nil {
			fmt.Println("💥 Error getting the coordinates from the address in CreateUser() : ", errGmaps)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		user.X = &coordinates.Lat
		user.Y = &coordinates.Lng

		request := fmt.Sprintf(`
			UPDATE public.user 
				SET X=%[1]s, Y=%[2]s, geom=ST_Buffer(st_transform(ST_SetSRID(ST_MakePoint(%[2]s, %[1]s), 4326), 2154), %[3]s, 'quad_segs=100')
			WHERE PhoneNumber='%[4]s'`,
			strconv.FormatFloat(*user.X, 'f', -1, 64),
			strconv.FormatFloat(*user.Y, 'f', -1, 64),
			strconv.FormatFloat(*user.Radius, 'f', -1, 64),
			user.PhoneNumber,
		)

		rows, err := db.Query(request)
		if err != nil {
			fmt.Println("💥 Error executing the request in CreateUser() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		// Close the rows
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				fmt.Println("💥 Error closing the rows in CreateUser()")
			}
		}(rows)
	} else {
		fmt.Println("=> Address is nil or not provided")
	}

	// Get the role name
	stmt, err = db.Prepare(`SELECT label FROM "role" WHERE IdRole = $1`)
	if err != nil {
		fmt.Println("💥 Error preparing the request in LoginUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in LoginUser()")
			return
		}
	}(stmt)

	var role string
	err = stmt.QueryRow(user.IdRole).Scan(&role)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	// Generate JWT token
	token, err := GenerateJWT(user.PhoneNumber, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate JWT token",
		})
	}

	// Return the token in the response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"Message": "User successfully created",
		"token":   token,
	})
}

func LoginUser(c *fiber.Ctx) error {
	email := c.Query("email")
	phoneNumber := c.Query("phone_number")
	password := c.Query("password")

	var query string
	var arg interface{}

	if email != "" {
		query = `SELECT Password FROM "user" WHERE Email = $1`
		arg = email
	} else if phoneNumber != "" {
		query = `SELECT Password FROM "user" WHERE PhoneNumber = $1`
		arg = phoneNumber
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Please provide an email or a phone number",
		})
	}

	stmt, err := db.Prepare(query)
	if err != nil {
		fmt.Println("💥 Error preparing the request in LoginUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in LoginUser()")
			return
		}
	}(stmt)

	var user User
	err = stmt.QueryRow(arg).Scan(&user.Password)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Vérification du mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid password",
		})
	}

	// Select idrole and phone number
	stmt, err = db.Prepare(`SELECT IdRole, PhoneNumber FROM "user" WHERE Email = $1 OR PhoneNumber = $2`)
	if err != nil {
		fmt.Println("💥 Error preparing the request in LoginUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in LoginUser()")
			return
		}
	}(stmt)

	err = stmt.QueryRow(email, phoneNumber).Scan(&user.IdRole, &user.PhoneNumber)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Get the role name
	stmt, err = db.Prepare(`SELECT label FROM "role" WHERE IdRole = $1`)
	if err != nil {
		fmt.Println("💥 Error preparing the request in LoginUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in LoginUser()")
			return
		}
	}(stmt)

	var role string
	err = stmt.QueryRow(user.IdRole).Scan(&role)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	// Generate JWT token
	token, err := GenerateJWT(user.PhoneNumber, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate JWT token",
		})
	}

	// Return the token in the response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token": token,
	})
}

// GetUser récupère un utilisateur spécifique à partir de son ID, ou tous les utilisateurs s'il n'y a pas d'ID spécifié.
func GetUser(c *fiber.Ctx) error {
	id := c.Query("id")
	//phoneNumber := ""
	phoneNumber := c.Locals("user").(*CustomClaims).PhoneNumber
	// Si un ID est spécifié dans les paramètres de la requête, on récupère uniquement cet utilisateur spécifique.
	if id != "" {
		//TODO Fix bug avec l'id qui est mal formaté ++33612345678 si postman, +33612345678 sur Voyo = problème
		var user User
		//id = strings.ReplaceAll("+"+id, " ", "")
		//fmt.Println("ID: ", id)
		stmt, err := db.Prepare(`SELECT PhoneNumber, FirstName, LastName, Biography, ProfilePicture, Pricing FROM "user" WHERE PhoneNumber = $1`)
		if err != nil {
			fmt.Println("💥 Error preparing the request to get one user in GetUser() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		defer func(stmt *sql.Stmt) {
			err := stmt.Close()
			if err != nil {
				fmt.Println("💥 Error closing the statement in GetUser()")
				return
			}
		}(stmt)

		row := stmt.QueryRow(id)
		err = row.Scan(&user.PhoneNumber, &user.FirstName, &user.LastName, &user.Biography, &user.ProfilePicture, &user.Pricing)
		if err != nil {
			fmt.Println("💥 Error scanning the row in GetUser() : ", err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.JSON(user)
	} else if phoneNumber != "" {
		var user User
		//id = strings.ReplaceAll("+"+id, " ", "")
		stmt, err := db.Prepare(`SELECT FirstName, LastName, Email, Biography, ProfilePicture, Pricing, Radius, x, y FROM "user" WHERE PhoneNumber = $1`)
		if err != nil {
			fmt.Println("💥 Error preparing the request to get one user in GetUser() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		defer func(stmt *sql.Stmt) {
			err := stmt.Close()
			if err != nil {
				fmt.Println("💥 Error closing the statement in GetUser()")
				return
			}
		}(stmt)

		row := stmt.QueryRow(phoneNumber)
		err = row.Scan(&user.FirstName, &user.LastName, &user.Email, &user.Biography, &user.ProfilePicture, &user.Pricing, &user.Radius, &user.X, &user.Y)
		if err != nil {
			fmt.Println("💥 Error scanning the row in GetUser() : ", err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.JSON(user)
	} else {
		// Select all users
		rows, err := db.Query(`SELECT PhoneNumber, FirstName, LastName, Email, Biography, ProfilePicture, Pricing, Radius, x, y FROM "user"`)
		if err != nil {
			fmt.Println("💥 Error querying the database in GetUser() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})

		}

		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				fmt.Println("💥 Error closing the rows in GetUser()")
				return
			}
		}(rows)

		var users []User
		for rows.Next() {
			var user User
			err := rows.Scan(&user.PhoneNumber, &user.FirstName, &user.LastName, &user.Email, &user.Biography, &user.ProfilePicture, &user.Pricing, &user.Radius, &user.X, &user.Y)
			if err != nil {
				fmt.Println("💥 Error scanning the rows in GetUser() : ", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "An error has occurred, please try again later.",
				})
			}
			users = append(users, user)
		}

		return c.JSON(users)
	}

	//return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	//	"error": "Please provide an ID",
	//})
}

// UpdateUser met à jour un utilisateur existant dans la base de données.
func UpdateUser(c *fiber.Ctx) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
		fmt.Println("💥 Error parsing the body in UpdateUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	user.PhoneNumber = c.Locals("user").(*CustomClaims).PhoneNumber

	var updateQuery string
	var args []interface{}

	placeholderIndex := 1 // Start with placeholder index 1

	if user.FirstName != "" {
		updateQuery += fmt.Sprintf(`FirstName=$%d,`, placeholderIndex)
		args = append(args, user.FirstName)
		placeholderIndex++
	}

	if user.LastName != "" {
		updateQuery += fmt.Sprintf(`LastName=$%d,`, placeholderIndex)
		args = append(args, user.LastName)
		placeholderIndex++
	}

	if user.Email != "" {
		updateQuery += fmt.Sprintf(`Email=$%d,`, placeholderIndex)
		args = append(args, user.Email)
		placeholderIndex++
	}

	if user.Biography != nil {
		updateQuery += fmt.Sprintf(`Biography=$%d,`, placeholderIndex)
		args = append(args, user.Biography)
		placeholderIndex++
	}

	if user.ProfilePicture != nil {
		updateQuery += fmt.Sprintf(`ProfilePicture=$%d,`, placeholderIndex)
		args = append(args, user.ProfilePicture)
		placeholderIndex++
	}

	if user.Pricing != nil {
		updateQuery += fmt.Sprintf(`Pricing=$%d,`, placeholderIndex)
		args = append(args, user.Pricing)
		placeholderIndex++
	}

	if user.IdAddressGMap != nil {
		updateQuery += fmt.Sprintf(`IdAddressGMap=$%d,`, placeholderIndex)
		args = append(args, user.IdAddressGMap)
		placeholderIndex++
	}

	if user.Radius != nil {
		updateQuery += fmt.Sprintf(`Radius=$%d,`, placeholderIndex)
		args = append(args, user.Radius)
		placeholderIndex++
	}

	if user.X != nil {
		updateQuery += fmt.Sprintf(`X=$%d,`, placeholderIndex)
		args = append(args, user.X)
		placeholderIndex++
	}

	if user.Y != nil {
		updateQuery += fmt.Sprintf(`Y=$%d,`, placeholderIndex)
		args = append(args, user.Y)
		placeholderIndex++
	}

	if user.CniFront != nil {
		updateQuery += fmt.Sprintf(`CniFront=$%d,`, placeholderIndex)
		args = append(args, user.CniFront)
		placeholderIndex++
	}

	if user.CniBack != nil {
		updateQuery += fmt.Sprintf(`CniBack=$%d,`, placeholderIndex)
		args = append(args, user.CniBack)
		placeholderIndex++
	}

	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println("💥 Error hashing the password in UpdateUser() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		updateQuery += fmt.Sprintf(`Password=$%d, PasswordUpdatedAt=NOW(), `, placeholderIndex)
		args = append(args, hashedPassword)
		placeholderIndex++
	}

	// Remove the last comma
	updateQuery = updateQuery[:len(updateQuery)-1]

	// Prepare the request
	stmt, err := db.Prepare(fmt.Sprintf(`
		UPDATE "user"
		SET %s
		WHERE PhoneNumber=$%d
	`, updateQuery, placeholderIndex))
	if err != nil {
		fmt.Println("💥 Error preparing the request in UpdateUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in UpdateUser()")
			return
		}

	}(stmt)

	args = append(args, user.PhoneNumber)
	_, err = stmt.Exec(args...)
	if err != nil {
		fmt.Println("💥 Error executing the request in UpdateUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DeleteUser supprime un utilisateur de la base de données.
func DeleteUser(c *fiber.Ctx) error {
	id := c.Query("id")

	stmt, err := db.Prepare(`DELETE FROM "user" WHERE PhoneNumber=$1`)
	if err != nil {
		fmt.Println("💥 Error preparing the request in DeleteUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		fmt.Println("💥 Error executing the request in DeleteUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	// TODO: instead of removing the user, set all data to null or User deleted in the database to avoid errors on foreign keys

	return c.SendStatus(fiber.StatusNoContent)
}

func GetHomeStats(c *fiber.Ctx) error {
	// Get the number of users
	role := c.Locals("user").(*CustomClaims).Role

	if role == "PROSPECT" {
		// 0) Struct to store all data
		type HomeStats struct {
			ProgrammedVisits int `json:"programmed_visits"`
			UnreadMessages   int `json:"unread_messages"`
			VisitedDone      int `json:"visited_done"`
			WaitingReviews   int `json:"waiting_reviews"`
		}

		var homeStats HomeStats

		// 1) Programmed visits
		stmt, err := db.Prepare(`SELECT COUNT(*) FROM visit WHERE PhoneNumberProspect = $1 AND Status IN ('PROGRAMMED', 'ACCEPTED') AND StartTime > NOW()`)
		if err != nil {
			fmt.Println("💥 Error preparing the request in GetHomeStats() programmed visits: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		defer func(stmt *sql.Stmt) {
			err := stmt.Close()
			if err != nil {
				fmt.Println("💥 Error closing the statement in GetHomeStats() programmed visits")
				return
			}
		}(stmt)

		err = stmt.QueryRow(c.Locals("user").(*CustomClaims).PhoneNumber).Scan(&homeStats.ProgrammedVisits)
		if err != nil {
			fmt.Println("💥 Error executing the request in GetHomeStats() programmed visits : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		// 2) Unread messages
		homeStats.UnreadMessages = rand.Int() % 10 // TODO: get the real number by requesting the firebase database

		// 3) Visits done
		stmt, err = db.Prepare(`SELECT COUNT(*) FROM visit WHERE PhoneNumberProspect = $1 AND Status = 'DONE'`)
		if err != nil {
			fmt.Println("💥 Error preparing the request in GetHomeStats() Visits done : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		err = stmt.QueryRow(c.Locals("user").(*CustomClaims).PhoneNumber).Scan(&homeStats.VisitedDone)
		if err != nil {
			fmt.Println("💥 Error executing the request in GetHomeStats() Visits done : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		// 4) Waiting reviews
		stmt, err = db.Prepare(`SELECT COUNT(*) FROM visit WHERE PhoneNumberProspect = $1 AND Status = 'DONE' AND Note IS NULL OR  Note = 0.0`)
		if err != nil {
			fmt.Println("💥 Error preparing the request in GetHomeStats() Waiting reviews: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		err = stmt.QueryRow(c.Locals("user").(*CustomClaims).PhoneNumber).Scan(&homeStats.WaitingReviews)
		if err != nil {
			fmt.Println("💥 Error executing the request in GetHomeStats() Waiting reviews : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		return c.JSON(homeStats)
	} else if role == "VISITOR" {
		type HomeStats struct {
			UpcomingVisits  int `json:"upcoming_visits"`
			UnreadMessages  int `json:"unread_messages"`
			AwatingApproval int `json:"awaiting_approval"`
			WaitingReviews  int `json:"waiting_reviews"`
		}

		var homeStats HomeStats

		// 1) Upcoming visits
		stmt, err := db.Prepare(`SELECT COUNT(*) FROM visit WHERE PhoneNumberVisitor = $1 AND Status IN ('PROGRAMMED', 'ACCEPTED') AND StartTime > NOW()`)
		if err != nil {
			fmt.Println("💥 Error preparing the request in GetHomeStats() upcoming visits: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		defer func(stmt *sql.Stmt) {
			err := stmt.Close()
			if err != nil {
				fmt.Println("💥 Error closing the statement in GetHomeStats() upcoming visits")
				return
			}
		}(stmt)

		err = stmt.QueryRow(c.Locals("user").(*CustomClaims).PhoneNumber).Scan(&homeStats.UpcomingVisits)
		if err != nil {
			fmt.Println("💥 Error executing the request in GetHomeStats() upcoming visits : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		// 2) Unread messages
		homeStats.UnreadMessages = rand.Int() % 10 // TODO: get the real number by requesting the firebase database

		// 3) Awaiting approval
		stmt, err = db.Prepare(`SELECT COUNT(*) FROM visit WHERE PhoneNumberVisitor = $1 AND Status = 'PROGRAMMED'`)
		if err != nil {
			fmt.Println("💥 Error preparing the request in GetHomeStats() awaiting approval: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		err = stmt.QueryRow(c.Locals("user").(*CustomClaims).PhoneNumber).Scan(&homeStats.AwatingApproval)
		if err != nil {
			fmt.Println("💥 Error executing the request in GetHomeStats() awaiting approval : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		// 4) Waiting reviews
		stmt, err = db.Prepare(`SELECT COUNT(*) FROM visit WHERE PhoneNumberVisitor = $1 AND Status = 'DONE' AND Note IS NULL OR  Note = 0.0`)
		if err != nil {
			fmt.Println("💥 Error preparing the request in GetHomeStats() Waiting reviews: ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		err = stmt.QueryRow(c.Locals("user").(*CustomClaims).PhoneNumber).Scan(&homeStats.WaitingReviews)
		if err != nil {
			fmt.Println("💥 Error executing the request in GetHomeStats() Waiting reviews : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		return c.JSON(homeStats)
	}

	return c.SendStatus(fiber.StatusOK)
}

func SearchUsers(c *fiber.Ctx) error {
	query := strings.Replace(c.Query("q"), " ", "%", -1)

	// Prepare the request
	stmt, err := db.Prepare(`SELECT PhoneNumber, FirstName, LastName, Email, IdRole, Biography, ProfilePicture, Pricing, idaddressgmap, Radius, x, y, status, cniback, cnifront FROM "user" WHERE PhoneNumber LIKE $1 OR FirstName LIKE $1 OR LastName LIKE $1 OR Email LIKE $1`)
	if err != nil {
		fmt.Println("💥 Error preparing the request in SearchUsers() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in SearchUsers()")
			return
		}
	}(stmt)

	rows, err := stmt.Query("%" + query + "%")
	if err != nil {
		fmt.Println("💥 Error querying the database in SearchUsers() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println("💥 Error closing the rows in SearchUsers()")
			return
		}
	}(rows)

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.PhoneNumber, &user.FirstName, &user.LastName, &user.Email, &user.IdRole, &user.Biography, &user.ProfilePicture, &user.Pricing, &user.IdAddressGMap, &user.Radius, &user.X, &user.Y, &user.Status, &user.CniBack, &user.CniFront)
		if err != nil {
			fmt.Println("💥 Error scanning the rows in SearchUsers() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		users = append(users, user)

	}

	return c.JSON(users)
}

func UsersToBeValidated(c *fiber.Ctx) error {
	rows, err := db.Query(`SELECT 																				PhoneNumber, FirstName, LastName, Email, IdRole, Biography, ProfilePicture, Pricing, idaddressgmap, Radius, x, y, status, cniback, cnifront FROM "user" WHERE status = 'PENDING_VALIDATION'`)
	if err != nil {
		fmt.Println("💥 Error querying the database in UsersToBeValidated() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})

	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println("💥 Error closing the rows in UsersToBeValidated()")
			return
		}
	}(rows)

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.PhoneNumber, &user.FirstName, &user.LastName, &user.Email, &user.IdRole, &user.Biography, &user.ProfilePicture, &user.Pricing, &user.IdAddressGMap, &user.Radius, &user.X, &user.Y, &user.Status, &user.CniBack, &user.CniFront)
		if err != nil {
			fmt.Println("💥 Error scanning the rows in UsersToBeValidated() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		users = append(users, user)
	}

	return c.JSON(users)
}

func GetUserStatus(c *fiber.Ctx) error {
	phoneNumber := c.Locals("user").(*CustomClaims).PhoneNumber

	stmt, err := db.Prepare(`SELECT status FROM "user" WHERE PhoneNumber = $1`)
	if err != nil {
		fmt.Println("💥 Error preparing the request in GetUserStatus() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in GetUserStatus()")
			return
		}
	}(stmt)

	var status string
	err = stmt.QueryRow(phoneNumber).Scan(&status)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(fiber.Map{"status": status})
}

func AdminUpdateUser(c *fiber.Ctx) error {
	phoneNumber := c.Query("id")

	var user User
	if err := c.BodyParser(&user); err != nil {
		fmt.Println("💥 Error parsing the body in AdminUpdateUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	var updateQuery string
	var args []interface{}

	placeholderIndex := 1 // Start with placeholder index 1

	if user.PhoneNumber != "" {
		updateQuery += fmt.Sprintf(`PhoneNumber=$%d, `, placeholderIndex)
		args = append(args, user.PhoneNumber)
		placeholderIndex++
	}

	if user.FirstName != "" {
		updateQuery += fmt.Sprintf(`FirstName=$%d, `, placeholderIndex)
		args = append(args, user.FirstName)
		placeholderIndex++
	}

	if user.LastName != "" {
		updateQuery += fmt.Sprintf(`LastName=$%d, `, placeholderIndex)
		args = append(args, user.LastName)
		placeholderIndex++
	}

	if user.Email != "" {
		updateQuery += fmt.Sprintf(`Email=$%d, `, placeholderIndex)
		args = append(args, user.Email)
		placeholderIndex++
	}

	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println("💥 Error hashing the password in AdminUpdateUser() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		updateQuery += fmt.Sprintf(`Password=$%d, `, placeholderIndex)
		args = append(args, hashedPassword)
		placeholderIndex++
	}

	if user.IdRole != 0 {
		updateQuery += fmt.Sprintf(`IdRole=$%d, `, placeholderIndex)
		args = append(args, user.IdRole)
		placeholderIndex++
	}

	if user.Biography != nil {
		updateQuery += fmt.Sprintf(`Biography=$%d, `, placeholderIndex)
		args = append(args, user.Biography)
		placeholderIndex++
	}

	if user.ProfilePicture != nil {
		updateQuery += fmt.Sprintf(`ProfilePicture=$%d, `, placeholderIndex)
		args = append(args, user.ProfilePicture)
		placeholderIndex++
	}

	if user.Pricing != nil {
		updateQuery += fmt.Sprintf(`Pricing=$%d, `, placeholderIndex)
		args = append(args, user.Pricing)
		placeholderIndex++
	}

	if user.IdAddressGMap != nil {
		updateQuery += fmt.Sprintf(`IdAddressGMap=$%d, `, placeholderIndex)
		args = append(args, user.IdAddressGMap)
		placeholderIndex++
	}

	if user.Radius != nil {
		updateQuery += fmt.Sprintf(`Radius=$%d, `, placeholderIndex)
		args = append(args, user.Radius)
		placeholderIndex++
	}

	if user.X != nil {
		updateQuery += fmt.Sprintf(`X=$%d, `, placeholderIndex)
		args = append(args, user.X)
		placeholderIndex++
	}

	if user.Y != nil {
		updateQuery += fmt.Sprintf(`Y=$%d, `, placeholderIndex)
		args = append(args, user.Y)
		placeholderIndex++
	}

	if user.Status != nil {
		if *user.Status == "VALIDATED" {
			updateQuery += fmt.Sprintf(`Status=$%d, CniFront='', CniBack='', `, placeholderIndex)
		} else {
			updateQuery += fmt.Sprintf(`Status=$%d, `, placeholderIndex)
		}
		args = append(args, user.Status)
		placeholderIndex++
	}

	// Remove the last comma and space
	updateQuery = strings.TrimSuffix(updateQuery, ", ")

	// Prepare the request
	stmt, err := db.Prepare(fmt.Sprintf(`UPDATE "user" SET %s WHERE PhoneNumber=$%d`, updateQuery, placeholderIndex))
	if err != nil {
		fmt.Println("💥 Error preparing the request in AdminUpdateUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in AdminUpdateUser()")
			return
		}
	}(stmt)

	args = append(args, phoneNumber)
	_, err = stmt.Exec(args...)
	if err != nil {
		fmt.Println("💥 Error executing the request in AdminUpdateUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusOK)

}
