package routes

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// checks whether user is admin or owner
// otherwise return unauthorized error
func (checkOwner CheckIfOwner) checkIfAdmin(c *fiber.Ctx) error {
	//get jwt token
	user := c.Locals("user").(*jwt.Token)
	// get jwt claims
	claims := user.Claims.(jwt.MapClaims)

	check := "admin"
	if checkOwner {
		check = "owner"
	}

	// if user is not admin or owner return unauthorized
	if !claims[check].(bool) {
		return c.Status(fiber.StatusUnauthorized).JSON(
			&fiber.Map{"error": "unauthorized"})
	}

	return c.Next()
}

// only owner can promote user to admin
func (r *Repo) promoteAdmin(c *fiber.Ctx) error {
	// get email from form data
	email := strings.ToLower(c.FormValue("email"))
	// get admin action from form data
	admin := c.FormValue("admin")

	// validate email input
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			&fiber.Map{"error": "email param isn't provided"})
	}

	var (
		err error
		msg string
	)

	// perform promotion or demotion based on admin action
	if admin == "demote" {
		err = r.Queries.DemoteAdmin(r.Ctx, email)
		msg = "admin has been demoted to user"
	} else {
		err = r.Queries.PromoteAdmin(r.Ctx, email)
		msg = "user has been promoted to admin"
	}

	// handle any errors from the database operation
	if err != nil {
		log.Println("Error in changing admin status in PromoteAdmin or DemoteAdmin db function:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	// return success message
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{"message": msg})
}
