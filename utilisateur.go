package main

import (
	"github.com/gofiber/fiber/v2"
)

func GetUtilisateur(c *fiber.Ctx) error {
	id := c.Query("id")

	// Si un ID est spécifié dans les paramètres de la requête, on récupère uniquement cet utilisateur spécifique.
	if id != "" {
		var u Utilisateur
		stmt, err := db.Prepare("SELECT * FROM Utilisateur WHERE IdUtilisateur = $1")
		if err != nil {
			return err
		}
		defer stmt.Close()

		row := stmt.QueryRow(id)
		err = row.Scan(&u.IdUtilisateur, &u.FirstName, &u.LastName, &u.Email, &u.Adresse, &u.Ville, &u.CodePostal, &u.Tel, &u.Note, &u.Description, &u.Password, &u.IdRole)
		if err != nil {
			return err
		}
		return c.JSON(u)
	}

	// Si aucun ID n'est spécifié, on récupère tous les utilisateurs.
	rows, err := db.Query("SELECT * FROM Utilisateur")
	if err != nil {
		return err
	}
	defer rows.Close()

	var utilisateurs []Utilisateur
	for rows.Next() {
		var u Utilisateur
		err := rows.Scan(&u.IdUtilisateur, &u.FirstName, &u.LastName, &u.Email, &u.Adresse, &u.Ville, &u.CodePostal, &u.Tel, &u.Note, &u.Description, &u.Password, &u.IdRole)
		if err != nil {
			return err
		}
		utilisateurs = append(utilisateurs, u)
	}

	return c.JSON(utilisateurs)

}

func CreateUtilisateur(c *fiber.Ctx) error {
	var u Utilisateur
	if err := c.BodyParser(&u); err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO Utilisateur (FirstName, LastName, Email, Adresse, Ville, CodePostal, Tel, Note, Description, Password, IdRole) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.FirstName, u.LastName, u.Email, u.Adresse, u.Ville, u.CodePostal, u.Tel, u.Note, u.Description, u.Password, u.IdRole)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).SendString("Utilisateur créé avec succès")
}

func UpdateUtilisateur(c *fiber.Ctx) error {
	id := c.Query("id")
	var u Utilisateur
	if err := c.BodyParser(&u); err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE Utilisateur SET FirstName=$1, LastName=$2, Email=$3, Adresse=$4, Ville=$5, CodePostal=$6, Tel=$7, Note=$8, Description=$9, Password=$10, IdRole=$11 WHERE IdUtilisateur=$12")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.FirstName, u.LastName, u.Email, u.Adresse, u.Ville, u.CodePostal, u.Tel, u.Note, u.Description, u.Password, u.IdRole, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func DeleteUtilisateur(c *fiber.Ctx) error {
	id := c.Query("id")

	stmt, err := db.Prepare("DELETE FROM Utilisateur WHERE IdUtilisateur=$1")
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
