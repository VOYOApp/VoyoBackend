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
func getAddressFromGMapsID(googleMapsID string) (googleMapsResponse, error) {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?key=%s&place_id=%s", apiKey, googleMapsID)
	// Make HTTP request
	response, err := http.Get(url)
	if err != nil {
		return googleMapsResponse{}, err
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
		return googleMapsResponse{}, err
	}

	// Parse the JSON response
	var result googleMapsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return googleMapsResponse{}, err
	}

	// Check if the response status is OK
	if result.Status != "OK" || len(result.Results) == 0 {
		return googleMapsResponse{}, fmt.Errorf("Geocoding API request failed")
	}

	// Extract coordinates from the first result
	//address := result.Results[0].FormattedAddress
	return result, nil
}

func SearchUsers(c *fiber.Ctx) error {
	// Get the query parameter
	x := c.Query("x")
	y := c.Query("y")

	date := c.Query("date")
	idTypeRealEstate := c.Query("duration")

	// TODO: vérifier que l'utilisateur n'a pas déjà une visite de programmée sur le créneau indiqué
	request := fmt.Sprintf(`
		SELECT DISTINCT u.FirstName,
                UPPER(CONCAT(LEFT(u.LastName, 1), '.')) AS LastName,
                u.profilepicture,
                navg.avg                                AS NoteAvg,
                u.pricing,
                u.phonenumber,
			    COALESCE(u.radius, 0),
			    COALESCE(u.x, 0),
			    COALESCE(u.y, 0),
                CASE
                    WHEN
                        (ST_Distance(u.geom, st_transform(ST_SetSRID(ST_MakePoint(%[2]s, %[1]s), 4326), 2154)) /
                         1000)::numeric %% 100 >= 50
                        THEN CEIL(ST_Distance(u.geom, st_transform(ST_SetSRID(ST_MakePoint(%[2]s, %[1]s), 4326),
                                                                   2154)) / 1000 / 100.0) * 100
                    ELSE FLOOR(ST_Distance(u.geom,
                                           st_transform(ST_SetSRID(ST_MakePoint(%[2]s, %[1]s), 4326), 2154)) /
                               1000 / 100.0) * 100
                    END                                 AS rounded_distance
		FROM "user" u
		         JOIN public.availability a ON u.phonenumber = a.phonenumber
		         JOIN LATERAL (SELECT AVG(note) AS avg
		                       FROM public.visit
		                       WHERE phonenumbervisitor = u.phonenumber
		                         AND status = 'DONE'
		                         AND note != 0.0) AS navg ON TRUE
		WHERE u.idrole = 1
		  AND st_intersects(
		        u.geom,
		        st_transform(ST_SetSRID(ST_MakePoint(%[2]s, %[1]s), 4326), 2154))
		  AND ((
		           repeat = 'DAILY'
		               AND availability::timestamp <= '%[3]s'::timestamp
		               AND EXTRACT(HOUR FROM availability) <= EXTRACT(HOUR FROM '%[3]s'::timestamp)
		               AND EXTRACT(MINUTE FROM availability) <= EXTRACT(MINUTE FROM '%[3]s'::timestamp)
		               AND availability::time + duration::interval >= '%[3]s'::TIME +
		                                                              CAST((SELECT duration FROM typerealestate WHERE idtyperealestate = 4) AS INTERVAL)
		           ) OR (
		           repeat = 'WEEKLY'
		               AND availability::timestamp <= '%[3]s'::timestamp
		               AND EXTRACT(DOW FROM availability) = EXTRACT(DOW FROM '%[3]s'::timestamp)
		               AND EXTRACT(HOUR FROM availability) <= EXTRACT(HOUR FROM '%[3]s'::timestamp)
		               AND EXTRACT(MINUTE FROM availability) <= EXTRACT(MINUTE FROM '%[3]s'::timestamp)
		               AND availability::time + duration::interval >= '%[3]s'::TIME +
		                                                              CAST((SELECT duration FROM typerealestate WHERE idtyperealestate = 4) AS INTERVAL)
		           ) OR (
		           repeat = 'MONTHLY'
		               AND availability::timestamp <= '%[3]s'::timestamp
		               AND EXTRACT(DAY FROM availability) = EXTRACT(DAY FROM '%[3]s'::timestamp)
		               AND EXTRACT(HOUR FROM availability) <= EXTRACT(HOUR FROM '%[3]s'::timestamp)
		               AND EXTRACT(MINUTE FROM availability) <= EXTRACT(MINUTE FROM '%[3]s'::timestamp)
		               AND availability::time + duration::interval >= '%[3]s'::TIME +
		                                                              CAST((SELECT duration FROM typerealestate WHERE idtyperealestate = 4) AS INTERVAL)
		           ) OR (
		           repeat = 'YEARLY'
		               AND availability::timestamp <= '%[3]s'::timestamp
		               AND EXTRACT(YEAR FROM availability) <= EXTRACT(YEAR FROM '%[3]s'::timestamp)
		               AND EXTRACT(HOUR FROM availability) <= EXTRACT(HOUR FROM '%[3]s'::timestamp)
		               AND EXTRACT(MINUTE FROM availability) <= EXTRACT(MINUTE FROM '%[3]s'::timestamp)
		               AND availability::time + duration::interval >= '%[3]s'::TIME +
		                                                              CAST((SELECT duration FROM typerealestate WHERE idtyperealestate = 4) AS INTERVAL)
		           ))`,
		x, y, date, idTypeRealEstate)

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

	type SearchUser struct {
		FirstName       string          `json:"firstName"`
		LastName        string          `json:"lastName"`
		ProfilePicture  string          `json:"profilePicture"`
		NoteAvg         sql.NullFloat64 `json:"noteAvg"`
		Pricing         float64         `json:"pricing"`
		PhoneNumber     string          `json:"phoneNumber"`
		Radius          float64         `json:"radius"`
		X               float64         `json:"x"`
		Y               float64         `json:"y"`
		RoundedDistance float64         `json:"roundedDistance"`
	}

	var users []SearchUser
	for rows.Next() {
		var user SearchUser
		err := rows.Scan(&user.FirstName, &user.LastName, &user.ProfilePicture, &user.NoteAvg, &user.Pricing, &user.PhoneNumber, &user.Radius, &user.X, &user.Y, &user.RoundedDistance)
		if err != nil {
			fmt.Println("💥 Error scanning the rows in SearchUsers() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}
		users = append(users, user)
	}

	// If there is no user, return an empty array
	if len(users) == 0 {
		query2 := fmt.Sprintf(`
			SELECT DISTINCT u.FirstName,
			                UPPER(CONCAT(LEFT(u.LastName, 1), '.')) AS LastName,
			                u.profilepicture,
			                COALESCE(navg.avg, 0)                   AS NoteAvg,
			                u.pricing,
			                u.phonenumber,
			                COALESCE(u.radius, 0),
			                COALESCE(u.x, 0),
			                COALESCE(u.y, 0),
			                COALESCE(CASE
			                    WHEN
			                        (ST_Distance(u.geom, st_transform(
			                                ST_SetSRID(ST_MakePoint(%[2]s, %[1]s), 4326), 2154)) /
			                         1000)::numeric %% 100 >= 50
			                        THEN CEIL(ST_Distance(u.geom, st_transform(
			                            ST_SetSRID(ST_MakePoint(%[2]s, %[1]s), 4326),
			                            2154)) / 1000 / 100.0) * 100
			                    ELSE FLOOR(ST_Distance(u.geom,
			                                           st_transform(
			                                                   ST_SetSRID(ST_MakePoint(%[2]s, %[1]s), 4326),
			                                                   2154)) /
			                               1000 / 100.0) * 100
			                    END, 0)                                  AS rounded_distance
			FROM "user" u
			         JOIN public.availability a ON u.phonenumber = a.phonenumber
			         JOIN LATERAL (SELECT AVG(note) AS avg
			                       FROM public.visit
			                       WHERE phonenumbervisitor = u.phonenumber
			                         AND status = 'DONE'
			                         AND note != 0.0) AS navg ON TRUE
			WHERE u.idrole = 1
			ORDER BY rounded_distance ASC
			LIMIT 20`, x, y)

		rows2, err := db.Query(query2)
		if err != nil {
			fmt.Println("💥 Error querying the database in SearchUsers() : ", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "An error has occurred, please try again later.",
			})
		}

		defer func(rows2 *sql.Rows) {
			err := rows2.Close()
			if err != nil {
				fmt.Println("💥 Error closing the rows in SearchUsers() : ", err)
				return
			}
		}(rows2)

		for rows2.Next() {
			var user SearchUser
			err := rows2.Scan(&user.FirstName, &user.LastName, &user.ProfilePicture, &user.NoteAvg, &user.Pricing, &user.PhoneNumber, &user.Radius, &user.X, &user.Y, &user.RoundedDistance)
			if err != nil {
				fmt.Println("💥 Error scanning the rows in SearchUsers() : ", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "An error has occurred, please try again later.",
				})
			}
			users = append(users, user)
		}
	}

	return c.JSON(users)
}

func checkUserExists(phoneNumber string) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM \"user\" WHERE phonenumber = $1)", phoneNumber).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func checkUserRole(phoneNumber string, role string) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM \"user\" WHERE phonenumber = $1 AND idrole = (SELECT idrole FROM role WHERE label = $2))", phoneNumber, role).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}
