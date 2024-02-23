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

	app.Get("/status", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString("1.0.0")
	})

	app.Post("/otp", sendOTP)
	app.Post("/verifyOTP", verifyOTP)

	// Routes pour la table "Realestate"
	app.Get("/realestate", GetRealEstate)
	app.Post("/realestate", CreateRealEstate)
	app.Put("/realestate", UpdateRealEstate)
	app.Delete("/realestate", DeleteRealEstate)

	// Routes pour la table "TypeRealeState"
	app.Get("/typerealestate", GetTypeRealEstate)
	app.Post("/typerealestate", CreateTypeRealEstate)
	app.Put("/typerealestate", UpdateTypeRealEstate)
	app.Delete("/typerealestate", DeleteTypeRealEstate)

	// Routes pour la table "Availability"
	app.Get("/availability", GetAvailability)
	app.Post("/availability", CreateAvailability)
	app.Put("/availability", UpdateAvailability)
	app.Delete("/availability", DeleteAvailability)

	// Routes pour la table "Role"
	app.Get("/role", GetRole)
	app.Post("/role", CreateRole)
	app.Put("/role", UpdateRole)
	app.Delete("/role", DeleteRole)

	// Routes pour la table "User"
	app.Get("/user", GetUser)
	app.Get("/login", LoginUser)
	app.Post("/inscription", CreateUser)
	app.Put("/user", UpdateUser)
	app.Delete("/user", DeleteUser)

	// Routes pour la table "Visit"
	app.Get("/visit", GetVisit)
	app.Post("/visit", CreateVisit)
	app.Put("/visit", UpdateVisit)
	app.Delete("/visit", DeleteVisit)

	// Routes pour la table "Criteria"
	app.Get("/criteria", GetCriteria)
	app.Post("/criteria", CreateCriteria)
	app.Put("/criteria", UpdateCriteria)
	app.Delete("/criteria", DeleteCriteria)

	// Routes pour la table "linkCriteriaVisit"
	app.Get("/linkcriteriavisit", GetLinkCriteriaVisit)
	app.Post("/linkcriteriavisit", CreateLinkCriteriaVisit)
	app.Delete("/linkcriteriavisit", DeleteLinkCriteriaVisit)

	fmt.Printf("Server is running on :%d...\n", 3000)
	err := app.Listen(":3000")
	if err != nil {
		fmt.Println(err)
		return
	}
}
