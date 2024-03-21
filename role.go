package main

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
)

func GetRole(c *fiber.Ctx) error {
	id := c.Query("id")

	// Si un ID est spécifié dans les paramètres de la requête,
	// on récupère uniquement ce rôle spécifique.
	if id != "" {
		var r Role
		stmt, err := db.Prepare("SELECT * FROM role WHERE IdRole = $1")
		if err != nil {
			return err
		}
		defer stmt.Close()

		row := stmt.QueryRow(id)
		err = row.Scan(&r.IdRole, &r.Label)
		if err != nil {
			return err
		}
		return c.JSON(r)
	}

	// Si aucun ID n'est spécifié, on récupère tous les rôles.
	rows, err := db.Query("SELECT * FROM role")
	if err != nil {
		return err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	var roles []Role
	for rows.Next() {
		var r Role
		err := rows.Scan(&r.IdRole, &r.Label)
		if err != nil {
			return err
		}
		roles = append(roles, r)
	}

	return c.JSON(roles)
}

func CreateRole(c *fiber.Ctx) error {
	var r Role
	if err := c.BodyParser(&r); err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO role (Label) VALUES ($1)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(r.Label)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).SendString("Rôle créé avec succès")
}

func UpdateRole(c *fiber.Ctx) error {
	id := c.Query("id")
	var r Role
	if err := c.BodyParser(&r); err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE role SET Label=$1 WHERE IdRole=$2")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(r.Label, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func DeleteRole(c *fiber.Ctx) error {
	id := c.Query("id")

	stmt, err := db.Prepare("DELETE FROM role WHERE IdRole=$1")
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
