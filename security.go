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
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) <= 7 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No token provided"})
	}

	tokenString := authHeader[7:] // Remove "Bearer " prefix if included

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			fmt.Println("Invalid token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		fmt.Println("💥 Error parsing the token in VerifyJWT() : ", err)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bad request"})
	}

	// Check if the token is valid
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// Token is valid, you can access claims like claims.UserID, claims.Email, etc.
		// Check if the user still exists
		if !UserExists(claims.PhoneNumber) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
		}

		// Check if the user changed password after the token was issued
		if PasswordChanged(claims.PhoneNumber, claims.CreatedAt) {
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
	request := fmt.Sprintf(`
		SELECT PhoneNumber
		FROM "user"
		WHERE PhoneNumber = '%s'
	`, PhoneNumber)

	// Execute the request
	row := db.QueryRow(request)
	var requestedPhoneNumber string
	err := row.Scan(&requestedPhoneNumber)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}

		fmt.Println("💥 Error scanning the row in UserExists() : ", err)
		return false
	}

	return requestedPhoneNumber == PhoneNumber
}

// PasswordChanged is a placeholder function to check if a user changed their password (replace with your own logic)
func PasswordChanged(PhoneNumber string, Date int64) bool {
	// 1) Get the password change date from the database
	request := fmt.Sprintf(`
		SELECT passwordUpdatedAt
		FROM "user"
		WHERE PhoneNumber = '%s'
	`, PhoneNumber)

	// Execute the request
	row := db.QueryRow(request)
	var passwordUpdatedAt string
	err := row.Scan(&passwordUpdatedAt)

	if err != nil {
		fmt.Println("💥 Error scanning the row in PasswordChanged() : ", err)
		return false
	}

	sqlDate, err := time.Parse(time.RFC3339Nano, passwordUpdatedAt)
	if err != nil {
		fmt.Println("💥 Error parsing the date in PasswordChanged() : ", err)
		return false
	}

	// Convert SQL date to Unix timestamp
	sqlUnix := sqlDate.Unix()

	// 2) Compare the password change date with the token creation date and return true if the password was changed
	// before the token was issued, false otherwise
	// TODO: fix this comparison (sometimes it returns true when it should return false)
	if sqlUnix > Date {
		return false
	} else {
		return false
	}

	return false
}

func hasAuthorizedVisitAccess(phoneNumber string, idVisit string) bool {
	request := fmt.Sprintf(`
		SELECT idvisit
		FROM "visit"
		WHERE idvisit = '%[1]s' AND (phonenumberprospect = '%[2]s' OR phonenumbervisitor = '%[2]s')
	`, idVisit, phoneNumber)

	// Execute the request
	row := db.QueryRow(request)
	var requestedId string
	err := row.Scan(&requestedId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}

		fmt.Println("💥 Error scanning the row in hasAuthorizedVisitAccess() : ", err)
		return false
	}

	return requestedId == idVisit
}

func hasAuthorizedCriteriaAccess(phoneNumber string, idCriteria string) bool {
	request := fmt.Sprintf(`
		SELECT idCriteria
		FROM linkcriteriavisit 
		    JOIN visit ON linkcriteriavisit.idVisit = visit.idVisit
		WHERE idcriteria = '%[1]s' AND (phonenumberprospect = '%[2]s' OR phonenumbervisitor = '%[2]s')
	`, idCriteria, phoneNumber)

	// Execute the request
	row := db.QueryRow(request)
	var requestedId string
	err := row.Scan(&requestedId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}

		fmt.Println("💥 Error scanning the row in hasAuthorizedCriteriaAccess() : ", err)
		return false
	}

	return requestedId == idCriteria
}

func restrictTo(role ...string) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		claims := c.Locals("user").(*CustomClaims)
		for _, r := range role {
			if claims.Role == r {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
}
