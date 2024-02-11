package main

import (
	"github.com/gofiber/fiber/v2"
)

func GetBien(c *fiber.Ctx) error {
	id := c.Params("id")
	var b Bien

	row := db.QueryRow("SELECT * FROM Bien WHERE IdBien = $1", id)
	err := row.Scan(&b.IdBien, &b.CodePostal, &b.Ville, &b.Adresse, &b.Proprietaire, &b.Pays)
	if err != nil {
		return err
	}

	return c.JSON(b)
}

func GetAllBien(c *fiber.Ctx) error {
	rows, err := db.Query("SELECT * FROM Bien")
	if err != nil {
		return err
	}
	defer rows.Close()

	var biens []Bien
	for rows.Next() {
		var b Bien
		err := rows.Scan(&b.IdBien, &b.CodePostal, &b.Ville, &b.Adresse, &b.Proprietaire, &b.Pays)
		if err != nil {
			return err
		}
		biens = append(biens, b)
	}

	return c.JSON(biens)
}

func CreateBien(c *fiber.Ctx) error {
	var b Bien
	if err := c.BodyParser(&b); err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO Bien (CodePostal, Ville, Adresse, Proprietaire, Pays) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(b.CodePostal, b.Ville, b.Adresse, b.Proprietaire, b.Pays)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).SendString("Bien créé avec succès")
}

func UpdateBien(c *fiber.Ctx) error {
	id := c.Params("id")
	var b Bien
	if err := c.BodyParser(&b); err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE Bien SET CodePostal=$1, Ville=$2, Adresse=$3, Proprietaire=$4, Pays=$5 WHERE IdBien=$6")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(b.CodePostal, b.Ville, b.Adresse, b.Proprietaire, b.Pays, id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func DeleteBien(c *fiber.Ctx) error {
	id := c.Params("id")

	stmt, err := db.Prepare("DELETE FROM Bien WHERE IdBien=$1")
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
