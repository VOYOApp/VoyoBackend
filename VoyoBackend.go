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

	// Define root routes
	root := app.Group("/api")
	root.Get("/status", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	root.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString("1.0.0")
	})

	// Define routes for "Realestate"
	realestate := root.Group("/realestate")
	realestate.Get("/", GetRealEstate)
	realestate.Post("/", CreateRealEstate)
	realestate.Put("/", UpdateRealEstate)
	realestate.Delete("/", DeleteRealEstate)

	// Define routes for "TypeRealeState"
	typerealestate := root.Group("/typerealestate")
	typerealestate.Get("/", GetTypeRealEstate)
	typerealestate.Post("/", CreateTypeRealEstate)
	typerealestate.Put("/", UpdateTypeRealEstate)
	typerealestate.Delete("/", DeleteTypeRealEstate)

	// Define routes for "Availability"
	availability := root.Group("/availability")
	availability.Get("/", GetAvailability)
	availability.Post("/", CreateAvailability)
	availability.Put("/", UpdateAvailability)
	availability.Delete("/", DeleteAvailability)

	// Define routes for "Role"
	role := root.Group("/role")
	role.Get("/", GetRole)
	role.Post("/", CreateRole)
	role.Put("/", UpdateRole)
	role.Delete("/", DeleteRole)

	// Define routes for "User"
	user := root.Group("/user")
	user.Get("/", GetUser)
	user.Get("/login", LoginUser)
	user.Post("/", CreateUser)
	user.Put("/", UpdateUser)
	user.Delete("/", VerifyJWT, DeleteUser)

	// Define routes for "Visit"
	visit := root.Group("/visit")
	visit.Get("/", GetVisit)
	visit.Post("/", CreateVisit)
	visit.Put("/", UpdateVisit)
	visit.Delete("/", DeleteVisit)

	// Define routes for "Criteria"
	criteria := root.Group("/criteria")
	criteria.Get("/", GetCriteria)
	criteria.Post("/", CreateCriteria)
	criteria.Put("/", UpdateCriteria)
	criteria.Delete("/", DeleteCriteria)

	// Define routes for "linkCriteriaVisit"
	linkcriteriavisit := root.Group("/linkcriteriavisit")
	linkcriteriavisit.Get("/", GetLinkCriteriaVisit)
	linkcriteriavisit.Post("/", CreateLinkCriteriaVisit)
	linkcriteriavisit.Delete("/", DeleteLinkCriteriaVisit)

	// Security routes
	security := root.Group("/security")
	security.Get("/", VerifyJWT, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Start the server
	fmt.Printf("Server is running on :%d...\n", 3000)
	err := app.Listen(":3000")
	if err != nil {
		fmt.Println(err)
		return
	}
}
