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

	// Define routes for "TypeRealeState"
	typerealestate := root.Group("/typerealestate")
	typerealestate.Get("/", GetTypeRealEstate)       // TODO: To check
	typerealestate.Post("/", CreateTypeRealEstate)   // TODO: To check
	typerealestate.Put("/", UpdateTypeRealEstate)    // TODO: To check
	typerealestate.Delete("/", DeleteTypeRealEstate) // TODO: To check

	// Define routes for "Availability"
	availability := root.Group("/availability")
	availability.Get("/", GetAvailability)       // TODO: To check
	availability.Post("/", CreateAvailability)   // TODO: To check
	availability.Put("/", UpdateAvailability)    // TODO: To check
	availability.Delete("/", DeleteAvailability) // TODO: To check

	// Define routes for "Role"
	role := root.Group("/role")
	role.Get("/", GetRole)       // TODO: To check
	role.Post("/", CreateRole)   // TODO: To check
	role.Put("/", UpdateRole)    // TODO: To check
	role.Delete("/", DeleteRole) // TODO: To check

	// Define routes for "User"
	user := root.Group("/user")
	user.Get("/", GetUser)        // TODO: To check 1 and add verifyJWT
	user.Get("/login", LoginUser) // TODO: To check
	user.Post("/", CreateUser)    // TODO: To check
	user.Put("/", UpdateUser)     // TODO: To check
	user.Delete("/", DeleteUser)  // TODO: To check
	user.Get("/homeStats", VerifyJWT, GetHomeStats)

	// Define routes for "Visit"
	visit := root.Group("/visit")
	visit.Get("/", VerifyJWT, GetVisit)
	visit.Patch("/", VerifyJWT, UpdateVisit)
	visit.Post("/", VerifyJWT, CreateVisit)
	visit.Delete("/", DeleteVisit) // TODO: To check
	visit.Get("/homeList", VerifyJWT, GetVisitsList)

	// Define routes for "Criteria"
	criteria := root.Group("/criteria")
	criteria.Get("/", VerifyJWT, GetCriteria)
	criteria.Post("/", VerifyJWT, CreateCriteria)
	criteria.Patch("/", VerifyJWT, UpdateCriteria) // TODO: To check 2
	criteria.Delete("/", DeleteCriteria)           // TODO: To check

	// Define routes for "linkCriteriaVisit" TODO: maybe will be deleted because a criteria is linked to a visit when creating the visit
	linkcriteriavisit := root.Group("/linkcriteriavisit")
	linkcriteriavisit.Get("/", GetLinkCriteriaVisit)       // TODO: To check
	linkcriteriavisit.Post("/", CreateLinkCriteriaVisit)   // TODO: To check
	linkcriteriavisit.Delete("/", DeleteLinkCriteriaVisit) // TODO: To check

	root.Get("/search", VerifyJWT, SearchUsers)

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
