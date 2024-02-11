package main

import "github.com/gofiber/fiber/v2"

func GetTravail(c *fiber.Ctx) error {
	idUtilisateur := c.Params("idUtilisateur")
	idLieux := c.Params("idLieux")
	var t Travail

	row := db.QueryRow("SELECT * FROM Travail WHERE idUtilisateur = $1 AND idLieux = $2", idUtilisateur, idLieux)
	err := row.Scan(&t.IdUtilisateur, &t.IdLieux)
	if err != nil {
		return err
	}

	return c.JSON(t)
}

func GetAllTravaux(c *fiber.Ctx) error {
	rows, err := db.Query("SELECT * FROM Travail")
	if err != nil {
		return err
	}
	defer rows.Close()

	var travaux []Travail
	for rows.Next() {
		var t Travail
		err := rows.Scan(&t.IdUtilisateur, &t.IdLieux)
		if err != nil {
			return err
		}
		travaux = append(travaux, t)
	}

	return c.JSON(travaux)
}

func CreateTravail(c *fiber.Ctx) error {
	var t Travail
	if err := c.BodyParser(&t); err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO Travail (idUtilisateur, idLieux) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.IdUtilisateur, t.IdLieux)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).SendString("Travail créé avec succès")
}

func UpdateTravail(c *fiber.Ctx) error {
	idUtilisateur := c.Params("idUtilisateur")
	idLieux := c.Params("idLieux")
	var t Travail
	if err := c.BodyParser(&t); err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE Travail SET idUtilisateur=$1, idLieux=$2 WHERE idUtilisateur=$3 AND idLieux=$4")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.IdUtilisateur, t.IdLieux, idUtilisateur, idLieux)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func DeleteTravail(c *fiber.Ctx) error {
	idUtilisateur := c.Params("idUtilisateur")
	idLieux := c.Params("idLieux")

	stmt, err := db.Prepare("DELETE FROM Travail WHERE idUtilisateur=$1 AND idLieux=$2")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(idUtilisateur, idLieux)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
