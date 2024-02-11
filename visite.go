package main

import "github.com/gofiber/fiber/v2"

func GetVisite(c *fiber.Ctx) error {
	idUtilisateur := c.Params("idUtilisateur")
	idUtilisateur1 := c.Params("idUtilisateur1")
	idBien := c.Params("idBien")
	var v Visite

	row := db.QueryRow("SELECT * FROM Visite WHERE idUtilisateur = $1 AND idUtilisateur_1 = $2 AND idBien = $3", idUtilisateur, idUtilisateur1, idBien)
	err := row.Scan(&v.IdUtilisateur, &v.IdUtilisateur1, &v.IdBien, &v.Agence, &v.CodeVerification, &v.Horaire, &v.APayer, &v.Etat)
	if err != nil {
		return err
	}

	return c.JSON(v)
}

func GetAllVisites(c *fiber.Ctx) error {
	rows, err := db.Query("SELECT * FROM Visite")
	if err != nil {
		return err
	}
	defer rows.Close()

	var visites []Visite
	for rows.Next() {
		var v Visite
		err := rows.Scan(&v.IdUtilisateur, &v.IdUtilisateur1, &v.IdBien, &v.Agence, &v.CodeVerification, &v.Horaire, &v.APayer, &v.Etat)
		if err != nil {
			return err
		}
		visites = append(visites, v)
	}

	return c.JSON(visites)
}

func CreateVisite(c *fiber.Ctx) error {
	var v Visite
	if err := c.BodyParser(&v); err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO Visite (idUtilisateur, idUtilisateur_1, idBien, agence, codeVerification, horaire, aPayer, etat) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(v.IdUtilisateur, v.IdUtilisateur1, v.IdBien, v.Agence, v.CodeVerification, v.Horaire, v.APayer, v.Etat)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).SendString("Visite créée avec succès")
}

func UpdateVisite(c *fiber.Ctx) error {
	idUtilisateur := c.Params("idUtilisateur")
	idUtilisateur1 := c.Params("idUtilisateur1")
	idBien := c.Params("idBien")
	var v Visite
	if err := c.BodyParser(&v); err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE Visite SET agence=$1, codeVerification=$2, horaire=$3, aPayer=$4, etat=$5 WHERE idUtilisateur=$6 AND idUtilisateur_1=$7 AND idBien=$8")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(v.Agence, v.CodeVerification, v.Horaire, v.APayer, v.Etat, idUtilisateur, idUtilisateur1, idBien)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func DeleteVisite(c *fiber.Ctx) error {
	idUtilisateur := c.Params("idUtilisateur")
	idUtilisateur1 := c.Params("idUtilisateur1")
	idBien := c.Params("idBien")

	stmt, err := db.Prepare("DELETE FROM Visite WHERE idUtilisateur=$1 AND idUtilisateur_1=$2 AND idBien=$3")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(idUtilisateur, idUtilisateur1, idBien)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
