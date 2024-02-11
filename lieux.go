package main

import "github.com/gofiber/fiber/v2"

func GetLieux(c *fiber.Ctx) error {
	id := c.Query("id")

	// Si un ID est spécifié dans les paramètres de la requête,
	// on récupère uniquement ce lieu spécifique.
	if id != "" {
		var l Lieux
		stmt, err := db.Prepare("SELECT * FROM Lieux WHERE IdLieux = $1")
		if err != nil {
			return err
		}
		defer stmt.Close()

		row := stmt.QueryRow(id)
		err = row.Scan(&l.IdLieux, &l.Radius, &l.Adresse, &l.Ville, &l.CodePostal, &l.Pays)
		if err != nil {
			return err
		}
		return c.JSON(l)
	}

	// Si aucun ID n'est spécifié, on récupère tous les lieux.
	rows, err := db.Query("SELECT * FROM Lieux")
	if err != nil {
		return err
	}
	defer rows.Close()

	var lieux []Lieux
	for rows.Next() {
		var l Lieux
		err := rows.Scan(&l.IdLieux, &l.Radius, &l.Adresse, &l.Ville, &l.CodePostal, &l.Pays)
		if err != nil {
			return err
		}
		lieux = append(lieux, l)
	}

	return c.JSON(lieux)
}

func CreateLieux(c *fiber.Ctx) error {
	var l Lieux
	if err := c.BodyParser(&l); err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO Lieux (Radius, Adresse, Ville, CodePostal, Pays) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(l.Radius, l.Adresse, l.Ville, l.CodePostal, l.Pays)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).SendString("Lieu créé avec succès")
}

func UpdateLieux(c *fiber.Ctx) error {
	id := c.Params("id")
	var l Lieux
	if err := c.BodyParser(&l); err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE Lieux SET Radius=$1, Adresse=$2, Ville=$3, CodePostal=$4, Pays=$5 WHERE IdLieux=$6")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(l.Radius, l.Adresse, l.Ville, l.CodePostal, l.Pays, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func DeleteLieux(c *fiber.Ctx) error {
	id := c.Params("id")

	stmt, err := db.Prepare("DELETE FROM Lieux WHERE IdLieux=$1")
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
