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

	stmt, err := db.Prepare(`INSERT INTO "user" (PhoneNumber, FirstName, LastName, Email, Password, IdRole, Biography, ProfilePicture, Pricing, IdAddressGMap, Radius) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`)
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

	_, err = stmt.Exec(user.PhoneNumber, user.FirstName, user.LastName, user.Email, hashedPassword, user.IdRole, user.Biography, user.ProfilePicture, user.Pricing, user.IdAddressGMap, user.Radius)
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
	phoneNumber := c.Locals("user").(*CustomClaims).PhoneNumber
	// Si un ID est spécifié dans les paramètres de la requête, on récupère uniquement cet utilisateur spécifique.
	if id != "" {
		var user User
		id = strings.ReplaceAll("+"+id, " ", "")
		stmt, err := db.Prepare(`SELECT PhoneNumber, FirstName, LastName, Biography, ProfilePicture, Pricing FROM "user" WHERE PhoneNumber = $1`)
		if err != nil {
			fmt.Println("💥 Error preparing the request to get one user in GetUser() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		defer stmt.Close()

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
		defer stmt.Close()

		row := stmt.QueryRow(phoneNumber)
		err = row.Scan(&user.FirstName, &user.LastName, &user.Email, &user.Biography, &user.ProfilePicture, &user.Pricing, &user.Radius, &user.X, &user.Y)
		if err != nil {
			fmt.Println("💥 Error scanning the row in GetUser() : ", err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.JSON(user)
	}

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "Please provide an ID",
	})
}

// Si aucun ID n'est spécifié, on récupère tous les utilisateurs.
//rows, err := db.Query(`SELECT PhoneNumber, FirstName, LastName, Email, Password, IdRole, Biography, ProfilePicture, Pricing, IdAddressGMap, Radius FROM "user"`)
//if err != nil {
//	fmt.Println("💥 Error preparing the request in GetUser() : ", err)
//	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//		"error": "An error has occurred, please try again later.",
//	})
//}
//defer rows.Close()
//
//var users []User
//for rows.Next() {
//	var user User
//	err := rows.Scan(&user.PhoneNumber, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.IdRole, &user.Biography, &user.ProfilePicture, &user.Pricing, &user.IdAddressGMap, &user.Radius)
//	if err != nil {
//		fmt.Println("💥 Error executing the request in GetUser() : ", err)
//		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//			"error": "An error has occurred, please try again later.",
//		})
//	}
//	users = append(users, user)
//}

//return c.JSON(users)
//}

// UpdateUser met à jour un utilisateur existant dans la base de données.
func UpdateUser(c *fiber.Ctx) error {
	id := c.Query("id")

	var user User
	if err := c.BodyParser(&user); err != nil {
		fmt.Println("💥 Error parsing the body in UpdateUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	stmt, err := db.Prepare(`UPDATE "user" SET PhoneNumber=$1, FirstName=$2, LastName=$3, Email=$4, Password=$5, IdRole=$6, Biography=$7, ProfilePicture=$8, Pricing=$9, IdAddressGMap=$10, Radius=$11 WHERE PhoneNumber=$12`)
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

	_, err = stmt.Exec(user.PhoneNumber, user.FirstName, user.LastName, user.Email, user.Password, user.IdRole, user.Biography, user.ProfilePicture, user.Pricing, user.IdAddressGMap, user.Radius, id)
	if err != nil {
		fmt.Println("💥 Error executing the request in UpdateUser() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	return c.SendStatus(fiber.StatusOK)
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
