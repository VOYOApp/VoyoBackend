package main

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"golang.org/x/crypto/bcrypt"
	"strconv"
)

// CreateUser crée un nouvel utilisateur dans la base de données.
func CreateUser(c *fiber.Ctx) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
		fmt.Println(err)
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	stmt, err := db.Prepare(`INSERT INTO "user" (PhoneNumber, FirstName, LastName, Email, Password, IdRole, Biography, ProfilePicture, Pricing, IdAddressGMap, Radius) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.PhoneNumber, user.FirstName, user.LastName, user.Email, hashedPassword, user.IdRole, user.Biography, user.ProfilePicture, user.Pricing, user.IdAddressGMap, user.Radius)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Get the X and Y coordinates from the address using the google maps api https://maps.googleapis.com/maps/api/geocode/json?key=AIzaSyBznSC8S1mPU-GPjsxuagQqnNK3a8xVOl4&place_id=
	if user.IdAddressGMap != nil {
		// Dereference the pointer and get coordinates
		address := utils.Trim(*user.IdAddressGMap, ' ')
		coordinates, err := getCoordinatesFromAddress(address)
		if err != nil {
			fmt.Println(err)
			return err
		}

		user.X = &coordinates.Lat
		user.Y = &coordinates.Lng
	} else {
		fmt.Println("Address is nil or not provided")
	}

	fmt.Println("")

	request := fmt.Sprintf(`
	UPDATE public.user 
	SET X=%[1]s, Y=%[2]s, geom=ST_Buffer(ST_SetSRID(ST_MakePoint(%[1]s, %[2]s), %[3]s), 500) 
	WHERE PhoneNumber='%[4]s'`,
		strconv.FormatFloat(*user.X, 'f', -1, 64),
		strconv.FormatFloat(*user.Y, 'f', -1, 64),
		strconv.FormatFloat(*user.Radius, 'f', -1, 64),
		user.PhoneNumber,
	)

	rows, err := db.Query(request)
	if err != nil {
		fmt.Println("💥 Error executing the request on createListeSufs()")
		fmt.Println(err)
		return err
	}

	// Close the rows
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println("💥 Error closing the rows")
		}
	}(rows)

	fmt.Sprintf("User %s (%s %s) successfully created",
		user.PhoneNumber,
		user.FirstName,
		user.LastName)

	return c.Status(fiber.StatusCreated).SendString("User successfully created")
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
		return c.Status(fiber.StatusBadRequest).SendString("Please specify an email or phone number")
	}

	stmt, err := db.Prepare(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer stmt.Close()

	var user User
	err = stmt.QueryRow(arg).Scan(&user.Password)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Vérification du mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Wrong password")
	}

	return c.Status(fiber.StatusOK).SendString("Successful connection")
}

// GetUser récupère un utilisateur spécifique à partir de son ID, ou tous les utilisateurs s'il n'y a pas d'ID spécifié.
func GetUser(c *fiber.Ctx) error {
	id := c.Query("id")

	// Si un ID est spécifié dans les paramètres de la requête, on récupère uniquement cet utilisateur spécifique.
	if id != "" {
		var user User

		stmt, err := db.Prepare(`SELECT PhoneNumber, FirstName, LastName, Email, Password, IdRole, Biography, ProfilePicture, Pricing, IdAddressGMap, Radius FROM "user" WHERE PhoneNumber = $1`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		row := stmt.QueryRow(id)
		err = row.Scan(&user.PhoneNumber, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.IdRole, &user.Biography, &user.ProfilePicture, &user.Pricing, &user.IdAddressGMap, &user.Radius)
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("User not found")
		}
		return c.JSON(user)
	}

	// Si aucun ID n'est spécifié, on récupère tous les utilisateurs.
	rows, err := db.Query(`SELECT PhoneNumber, FirstName, LastName, Email, Password, IdRole, Biography, ProfilePicture, Pricing, IdAddressGMap, Radius FROM "user"`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.PhoneNumber, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.IdRole, &user.Biography, &user.ProfilePicture, &user.Pricing, &user.IdAddressGMap, &user.Radius)
		if err != nil {
			return err
		}
		users = append(users, user)
	}

	return c.JSON(users)
}

// UpdateUser met à jour un utilisateur existant dans la base de données.
func UpdateUser(c *fiber.Ctx) error {
	id := c.Query("id")

	var user User
	if err := c.BodyParser(&user); err != nil {
		return err
	}

	stmt, err := db.Prepare(`UPDATE "user" SET PhoneNumber=$1, FirstName=$2, LastName=$3, Email=$4, Password=$5, IdRole=$6, Biography=$7, ProfilePicture=$8, Pricing=$9, IdAddressGMap=$10, Radius=$11 WHERE PhoneNumber=$12`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.PhoneNumber, user.FirstName, user.LastName, user.Email, user.Password, user.IdRole, user.Biography, user.ProfilePicture, user.Pricing, user.IdAddressGMap, user.Radius, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteUser supprime un utilisateur de la base de données.
func DeleteUser(c *fiber.Ctx) error {
	id := c.Query("id")

	stmt, err := db.Prepare(`DELETE FROM "user" WHERE PhoneNumber=$1`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
