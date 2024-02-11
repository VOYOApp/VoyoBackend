package main

import "github.com/gofiber/fiber/v2"

func GetRole(c *fiber.Ctx) error {
	id := c.Params("id")
	var r Role

	row := db.QueryRow("SELECT * FROM Role WHERE IdRole = $1", id)
	err := row.Scan(&r.IdRole, &r.Libelle)
	if err != nil {
		return err
	}

	return c.JSON(r)
}

func GetAllRoles(c *fiber.Ctx) error {
	rows, err := db.Query("SELECT * FROM Role")
	if err != nil {
		return err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var r Role
		err := rows.Scan(&r.IdRole, &r.Libelle)
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

	stmt, err := db.Prepare("INSERT INTO Role (Libelle) VALUES ($1)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(r.Libelle)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).SendString("Rôle créé avec succès")
}

func UpdateRole(c *fiber.Ctx) error {
	id := c.Params("id")
	var r Role
	if err := c.BodyParser(&r); err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE Role SET Libelle=$1 WHERE IdRole=$2")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(r.Libelle, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func DeleteRole(c *fiber.Ctx) error {
	id := c.Params("id")

	stmt, err := db.Prepare("DELETE FROM Role WHERE IdRole=$1")
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
