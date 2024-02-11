package main

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"os"
)

var db *sql.DB

// EXAMPLE database := "postgresql://username:password@localhost:5432/database"
//const (
//	host     = "gigondas"
//	port     = 5432
//	user     = "graillet"
//	password = "graillet"
//	dbname   = "voyo_db"
//)

func init() {
	err := godotenv.Load()

	// Connect to the database
	db, err = sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
		),
	)

	// If there is an error connecting to the database, exit the program
	if err != nil {
		fmt.Println("💥 Error connecting to the database")
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println("Connected to the database")
	}

}

func main() {
	app := fiber.New()

	//app.Use(cors.New(cors.Config{
	//	AllowOrigins: "*",
	//	AllowMethods: "GET",
	//	AllowHeaders: "Content-Type",
	//}))

	app.Get("/users", getUsers)
	app.Get("/users/:id", getUser)

	app.Get("/status", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString("1.0.0")
	})

	app.Post("/otp", sendOTP)
	app.Post("/verifyOTP", verifyOTP)

	// Routes pour la table "Bien"
	app.Get("/biens", GetBien)
	app.Post("/biens", CreateBien)
	app.Put("/biens/:id", UpdateBien)
	app.Delete("/biens/:id", DeleteBien)

	// Routes pour la table "Role"
	app.Get("/roles", GetRole)
	app.Post("/roles", CreateRole)
	app.Put("/roles/:id", UpdateRole)
	app.Delete("/roles/:id", DeleteRole)

	// Routes pour la table "Lieux"
	app.Get("/lieux", GetLieux)
	app.Post("/lieux", CreateLieux)
	app.Put("/lieux/:id", UpdateLieux)
	app.Delete("/lieux/:id", DeleteLieux)

	// Routes pour la table "Utilisateur"
	app.Get("/utilisateurs", GetUtilisateur)
	app.Post("/utilisateurs", CreateUtilisateur)
	app.Put("/utilisateurs/:id", UpdateUtilisateur)
	app.Delete("/utilisateurs/:id", DeleteUtilisateur)

	// Routes pour la table "Calendrier"
	app.Get("/calendriers", GetCalendrier)
	app.Post("/calendriers", CreateCalendrier)
	app.Put("/calendriers/:id", UpdateCalendrier)
	app.Delete("/calendriers/:id", DeleteCalendrier)

	// Routes pour la table "Visite"
	app.Get("/visites", GetVisite)
	app.Post("/visites", CreateVisite)
	app.Put("/visites/:idUtilisateur/:idUtilisateur1/:idBien", UpdateVisite)
	app.Delete("/visites/:idUtilisateur/:idUtilisateur1/:idBien", DeleteVisite)

	// Routes pour la table "Travail"
	app.Get("/travaux", GetTravail)
	app.Post("/travaux", CreateTravail)
	app.Put("/travaux/:idUtilisateur/:idLieux", UpdateTravail)
	app.Delete("/travaux/:idUtilisateur/:idLieux", DeleteTravail)

	fmt.Printf("Server is running on :%d...\n", 3000)
	err := app.Listen(":3000")
	if err != nil {
		fmt.Println(err)
		return
	}
}
