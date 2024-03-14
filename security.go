// security.go
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

var jwtSecret = []byte(os.Getenv("JWT_KEY")) // Change this to a strong secret key

// CustomClaims represents the custom claims for the JWT.
type CustomClaims struct {
	jwt.StandardClaims
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
	CreatedAt   int64  `json:"created_at"`
}

// GenerateJWT generates a JWT token for a user.
func GenerateJWT(PhoneNumber string, role string) (string, error) {
	// Set expiration time for the token (e.g., 1 day)
	expirationTime := time.Now().Add(240000 * time.Hour)

	// Create JWT claims
	claims := &CustomClaims{
		PhoneNumber: PhoneNumber,
		Role:        role,
		CreatedAt:   time.Now().Unix(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Create token with claims and secret key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// VerifyJWT middleware checks the validity of the JWT, user existence, and password changes.
func VerifyJWT(c *fiber.Ctx) error {
	// Get the token from the Authorization header
	tokenString := c.Get("Authorization")[7:] // Assuming "Bearer " is included in the header

	fmt.Println(tokenString)

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	fmt.Println(token)

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			fmt.Println("Invalid token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}
		fmt.Println("Bad request")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bad request"})
	}

	// Check if the token is valid
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {

		fmt.Println(claims)
		// Token is valid, you can access claims like claims.UserID, claims.Email, etc.
		// Check if the user still exists (replace with your own user existence check logic)

		fmt.Println(claims.PhoneNumber)

		if !UserExists(claims.PhoneNumber) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
		}

		fmt.Println(claims.PhoneNumber)

		// Check if the user changed password after the token was issued (replace with your own logic)
		if PasswordChanged(claims.PhoneNumber) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Password changed"})
		}

		// User and token are valid, you can store claims in locals for further use
		c.Locals("user", claims)
		return c.Next()
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
}

// UserExists is a placeholder function to check if a user exists (replace with your own logic)
func UserExists(PhoneNumber string) bool {
	// Prepare the request
	stmt, err := db.Prepare(`SELECT * FROM "user" WHERE PhoneNumber = $1`)

	if err != nil {
		fmt.Println("💥 Error preparing the request in UserExists() : ", err)
		return false
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			fmt.Println("💥 Error closing the statement in UserExists()")
			return
		}
	}(stmt)

	// Execute the request
	row := stmt.QueryRow(PhoneNumber)

	// Check if the user exists
	err = row.Scan()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}

		fmt.Println("💥 Error scanning the row in UserExists() : ", err)
		return false
	}

	// User exists
	return true
}

// PasswordChanged is a placeholder function to check if a user changed their password (replace with your own logic)
func PasswordChanged(PhoneNumber string) bool {
	// Implement your password change check logic here
	// Example: return true if password changed, false otherwise
	return false
}
