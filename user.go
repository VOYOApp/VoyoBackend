package main

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
)

// CreateUser crée un nouvel utilisateur dans la base de données.
func CreateUser(c *fiber.Ctx) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO user (PhoneNumber, FirstName, LastName, Email, Description, Password, IdRole, Biography, ProfilePicture, Pricing, IdAddressGMap, Radius) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.PhoneNumber, user.FirstName, user.LastName, user.Email, user.Description, user.Password, user.IdRole, user.Biography, user.ProfilePicture, user.Pricing, user.IdAddressGMap, user.Radius)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).SendString("Utilisateur créé avec succès")
}

// GetUser récupère un utilisateur spécifique à partir de son ID.
func GetUser(c *fiber.Ctx) error {
	id := c.Query("id")

	var user User
	err := db.QueryRow("SELECT PhoneNumber, FirstName, LastName, Email, Description, Password, IdRole, Biography, ProfilePicture, Pricing, IdAddressGMap, Radius FROM user WHERE ID = $1", id).Scan(&user.PhoneNumber, &user.FirstName, &user.LastName, &user.Email, &user.Description, &user.Password, &user.IdRole, &user.Biography, &user.ProfilePicture, &user.Pricing, &user.IdAddressGMap, &user.Radius)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Utilisateur non trouvé")
		}
		return err
	}

	return c.JSON(user)
}

// UpdateUser met à jour un utilisateur existant dans la base de données.
func UpdateUser(c *fiber.Ctx) error {
	id := c.Query("id")

	var user User
	if err := c.BodyParser(&user); err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE user SET PhoneNumber=$1, FirstName=$2, LastName=$3, Email=$4, Description=$5, Password=$6, IdRole=$7, Biography=$8, ProfilePicture=$9, Pricing=$10, IdAddressGMap=$11, Radius=$12 WHERE ID=$13")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.PhoneNumber, user.FirstName, user.LastName, user.Email, user.Description, user.Password, user.IdRole, user.Biography, user.ProfilePicture, user.Pricing, user.IdAddressGMap, user.Radius, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteUser supprime un utilisateur de la base de données.
func DeleteUser(c *fiber.Ctx) error {
	id := c.Query("id")

	stmt, err := db.Prepare("DELETE FROM user WHERE ID=$1")
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
