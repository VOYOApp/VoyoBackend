package main

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser crée un nouvel utilisateur dans la base de données.
func CreateUser(c *fiber.Ctx) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
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

	return c.Status(fiber.StatusCreated).SendString("Utilisateur créé avec succès")
}

func LoginUserWithPhoneNumber(c *fiber.Ctx) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
		return err
	}

	stmt, err := db.Prepare(`SELECT Password FROM "user" WHERE PhoneNumber = $1`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Password)
	if err != nil {
		return err
	}

	// Vérification du mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(user.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Mot de passe incorrect")
	}

	return c.SendStatus(fiber.StatusOK)
}

//func LoginUserWithEmail(c *fiber.Ctx) error {
//
//}

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
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).SendString("Utilisateur non trouvé")
			}
			return err
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
