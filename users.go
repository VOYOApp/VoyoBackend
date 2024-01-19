package main

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func getUsers(c *fiber.Ctx) error {

	// 1) Prepare the SQL query
	request := fmt.Sprintf("SELECT * FROM users")

	// 2) Execute the query
	rows, err := db.Query(request)
	if err != nil {
		fmt.Println("💥 Error executing the request on getUsers")
		fmt.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// 2.2) Close the rows (mandatory to avoid memory leaks)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println("💥 Error closing the rows")
		}
	}(rows)

	// 3) Loop over the result
	// 3.1) Create a slice of structs
	var users []User // Type of the slice (=array) is User (defined in the structs.go file)

	for rows.Next() {
		// 3.2) Create a struct
		var user User // Type of the struct is User (defined in the structs.go file)

		// 3.3) Scan the result into the struct
		err = rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			fmt.Println("💥 Error scanning the result into the struct")
			fmt.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		// 3.4) Append the struct to the slice
		users = append(users, user)
	}

	// 4) Return the struct as JSON
	return c.JSON(users)
}

func getUser(c *fiber.Ctx) error {

	// 1) Get the id from the URL
	id := c.Params("id")

	// 2) Prepare the SQL query
	request := fmt.Sprintf("SELECT * FROM users WHERE id = %s", id)

	// 3) Execute the query
	row := db.QueryRow(request)

	// 4) Create a struct
	var user User // Type of the struct is User (defined in the structs.go file)

	// 5) Scan the result into the struct
	err := row.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		fmt.Println("💥 Error scanning the result into the struct")
		fmt.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// 6) Return the struct as JSON
	return c.JSON(user)
}
