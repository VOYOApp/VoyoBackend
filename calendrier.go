package main

import "github.com/gofiber/fiber/v2"

func GetCalendrier(c *fiber.Ctx) error {
	id := c.Params("id")
	var cal Calendrier

	row := db.QueryRow("SELECT * FROM Calendrier WHERE IdCalendrier = $1", id)
	err := row.Scan(&cal.IdCalendrier, &cal.IdUtilisateur, &cal.Disponibilite, &cal.Temps)
	if err != nil {
		return err
	}

	return c.JSON(cal)
}

func GetAllCalendriers(c *fiber.Ctx) error {
	rows, err := db.Query("SELECT * FROM Calendrier")
	if err != nil {
		return err
	}
	defer rows.Close()

	var calendriers []Calendrier
	for rows.Next() {
		var cal Calendrier
		err := rows.Scan(&cal.IdCalendrier, &cal.IdUtilisateur, &cal.Disponibilite, &cal.Temps)
		if err != nil {
			return err
		}
		calendriers = append(calendriers, cal)
	}

	return c.JSON(calendriers)
}

func CreateCalendrier(c *fiber.Ctx) error {
	var cal Calendrier
	if err := c.BodyParser(&cal); err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO Calendrier (IdUtilisateur, Disponibilite, Temps) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(cal.IdUtilisateur, cal.Disponibilite, cal.Temps)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).SendString("Calendrier créé avec succès")
}

func UpdateCalendrier(c *fiber.Ctx) error {
	id := c.Params("id")
	var cal Calendrier
	if err := c.BodyParser(&cal); err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE Calendrier SET IdUtilisateur=$1, Disponibilite=$2, Temps=$3 WHERE IdCalendrier=$4")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(cal.IdUtilisateur, cal.Disponibilite, cal.Temps, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func DeleteCalendrier(c *fiber.Ctx) error {
	id := c.Params("id")

	stmt, err := db.Prepare("DELETE FROM Calendrier WHERE IdCalendrier=$1")
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
