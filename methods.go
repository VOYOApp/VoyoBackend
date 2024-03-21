package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"os"
)

// Function to get coordinates from address using the Google Maps Geocoding API
func getCoordinatesFromAddress(address string) (googleMapsCoordinates, error) {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?key=%s&place_id=%s", apiKey, address)
	// Make HTTP request
	response, err := http.Get(url)
	if err != nil {
		return googleMapsCoordinates{}, err
	}
	defer response.Body.Close()

	// Read and parse the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return googleMapsCoordinates{}, err
	}

	// Parse the JSON response
	var result googleMapsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return googleMapsCoordinates{}, err
	}

	// Check if the response status is OK
	if result.Status != "OK" || len(result.Results) == 0 {
		return googleMapsCoordinates{}, fmt.Errorf("Geocoding API request failed")
	}

	// Extract coordinates from the first result
	coordinates := result.Results[0].Geometry.Location
	return coordinates, nil
}

// Function to get coordinates from address using the Google Maps Geocoding API
func getAddressFromGMapsID(googleMapsID string) (string, error) {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?key=%s&place_id=%s", apiKey, googleMapsID)
	// Make HTTP request
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("💥 Error closing the response body in getAddressFromGMapsID() : ", err)
			return
		}
	}(response.Body)

	// Read and parse the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	// Parse the JSON response
	var result googleMapsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Check if the response status is OK
	if result.Status != "OK" || len(result.Results) == 0 {
		return "", fmt.Errorf("Geocoding API request failed")
	}

	// Extract coordinates from the first result
	address := result.Results[0].FormattedAddress
	return address, nil
}

func SearchUsers(c *fiber.Ctx) error {
	// Get the query parameter
	x := c.Query("x")
	y := c.Query("y")
	date := c.Query("date")
	duration := c.Query("duration")

	request := fmt.Sprintf(`
		SELECT DISTINCT u.FirstName,
                UPPER(CONCAT(LEFT(u.LastName, 1), '.')) AS LastName,
                u.profilepicture,
                navg.avg                                AS NoteAvg,
                u.pricing,
                u.phonenumber,
                CASE
                    WHEN
                        (ST_Distance(u.geom, st_transform(ST_SetSRID(ST_MakePoint(%[2], %[1]), 4326), 2154)) /
                         1000)::numeric % 100 >= 50
                        THEN CEIL(ST_Distance(u.geom, st_transform(ST_SetSRID(ST_MakePoint(%[2], %[1]), 4326),
                                                                   2154)) / 1000 / 100.0) * 100
                    ELSE FLOOR(ST_Distance(u.geom,
                                           st_transform(ST_SetSRID(ST_MakePoint(%[2], %[1]), 4326), 2154)) /
                               1000 / 100.0) * 100
                    END                                 AS rounded_distance

		FROM "user" u
		         JOIN public.availability a ON u.phonenumber = a.phonenumber
		         JOIN LATERAL (SELECT AVG(note) AS avg
		                       FROM public.visit
		                       WHERE phonenumbervisitor = u.phonenumber
		                         AND status = 'DONE'
		                         AND note != 0.0) AS navg ON TRUE
		WHERE st_intersects(
		        u.geom,
		        st_transform(ST_SetSRID(ST_MakePoint(%[2], %[1]), 4326), 2154))
		  AND ((
		           repeat = 'DAILY'
		               AND availability::timestamp <= '%[3]'::timestamp
		               AND EXTRACT(HOUR FROM availability) <= EXTRACT(HOUR FROM '%[3]'::timestamp)
		               AND EXTRACT(MINUTE FROM availability) <= EXTRACT(MINUTE FROM '%[3]'::timestamp)
		               AND availability::time + duration::interval >= '%[4]'::time
		           ) OR (
		           repeat = 'WEEKLY'
		               AND availability::timestamp <= '%[3]'::timestamp
		               AND EXTRACT(DOW FROM availability) = EXTRACT(DOW FROM '%[3]'::timestamp)
		               AND EXTRACT(HOUR FROM availability) <= EXTRACT(HOUR FROM '%[3]'::timestamp)
		               AND EXTRACT(MINUTE FROM availability) <= EXTRACT(MINUTE FROM '%[3]'::timestamp)
		               AND availability::time + duration::interval >= '%[4]'::time
		           ) OR (
		            repeat = 'MONTHLY'
		                AND availability::timestamp <= '%[3]'::timestamp
		                AND EXTRACT(DAY FROM availability) = EXTRACT(DAY FROM '%[3]'::timestamp)
		                AND EXTRACT(HOUR FROM availability) <= EXTRACT(HOUR FROM '%[3]'::timestamp)
		                AND EXTRACT(MINUTE FROM availability) <= EXTRACT(MINUTE FROM '%[3]'::timestamp)
		                AND availability::time + duration::interval >= '%[4]'::time
		           ) OR (
		           repeat = 'YEARLY'
		               AND availability::timestamp <= '%[3]'::timestamp
		               AND EXTRACT(YEAR FROM availability) <= EXTRACT(YEAR FROM '%[3]'::timestamp)
		               AND EXTRACT(HOUR FROM availability) <= EXTRACT(HOUR FROM '%[3]'::timestamp)
		               AND EXTRACT(MINUTE FROM availability) <= EXTRACT(MINUTE FROM '%[3]'::timestamp)
		               AND availability::time + duration::interval >= '%[4]'::time
		           ))`,
		x, y, date, duration)

	rows, err := db.Query(request)
	if err != nil {
		fmt.Println("💥 Error querying the database in SearchUsers() : ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "An error has occurred, please try again later.",
		})
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println("💥 Error closing the rows in SearchUsers() : ", err)
			return
		}
	}(rows)

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.PhoneNumber, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.IdRole, &u.Biography, &u.ProfilePicture, &u.Pricing, &u.IdAddressGMap, &u.Radius, &u.X, &u.Y, &u.Geom)
		if err != nil {
			fmt.Println("💥 Error scanning the rows in SearchUsers() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		users = append(users, u)
	}

	return c.JSON(users)
}
