package routes

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ItsYourAbhi/goTrip/db"
)

// getDestinations retrieves all destinations from the database
func (r *Repo) ListDestinations(c *fiber.Ctx) error {
	// get destinations from db
	destinations, err := r.Queries.ListDestinations(r.Ctx)

	if err != nil {
		log.Println("Error in getting destinations in GetDestinations db function:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	if len(destinations) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(
			&fiber.Map{"error": "no destinations found"})
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{"data": &destinations})
}

// getDestination retrieves a single destination by ID
func (r *Repo) getDestination(c *fiber.Ctx) error {
	uuid, err := uuid.Parse(c.Params("id"))

	if err != nil {
		if strings.HasPrefix(err.Error(), "invalid UUID") {
			return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidID)
		}

		log.Println("Error in parsing uuid:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	id := pgtype.UUID{
		Bytes: uuid,
		Valid: true,
	}

	// get destination using id
	destination, err := r.Queries.GetDestination(r.Ctx, id)

	if err != nil {
		// if destination id not in db return error invalid id
		if err.Error() == "no rows in result set" {
			return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidID)
		}

		log.Println("Error in getting destination in GetDestination db function:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{"data": &destination})
}

// createDestination adds a new destination to the database
func (r *Repo) createDestination(c *fiber.Ctx) error {
	// Extract name, description, and password from form data
	id := c.FormValue("id")
	name := c.FormValue("name")
	description := c.FormValue("description")
	attraction := c.FormValue("attraction")
	picUrl := c.FormValue("pic_url")

	// if any of the form data is missing return error
	if name == "" || description == "" || attraction == "" || picUrl == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiberUndefinedParamError)
	}

	Uuid := uuid.New()
	if id != "" {
		var err error
		Uuid, err = uuid.Parse(id)

		if err != nil {
			// return error if id is invalid
			if strings.HasPrefix(err.Error(), "invalid UUID") {
				return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidDestinationID)
			}

			log.Println("Error in parsing uuid:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
		}
	}

	if len(name) > 128 {
		return c.Status(fiber.StatusBadRequest).JSON(fiberNameTooLong128)
	}

	destination := db.CreateDestinationParams{
		ID: pgtype.UUID{
			Bytes: Uuid,
			Valid: true,
		},
		Name:        name,
		Description: description,
		Attraction:  attraction,
		PicUrl:      picUrl,
	}

	err := r.Queries.CreateDestination(r.Ctx, destination)

	if err != nil {
		log.Println("Error in creating destination in Createdestination db function:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	return c.Status(fiber.StatusCreated).JSON(&fiber.Map{
		"message": "destination has been added"})
}

// updateDestination modifies an existing destination in the database
func (r *Repo) updateDestination(c *fiber.Ctx) error {
	// Extract values from form data
	id := c.FormValue("id")
	name := c.FormValue("name")
	description := c.FormValue("description")
	attraction := c.FormValue("attraction")
	picUrl := c.FormValue("pic_url")

	// if any of the form data is missing return error
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiberUndefinedParamError)
	}

	if name == "" && description == "" && attraction == "" && picUrl == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiberUndefinedParamError)
	}

	if len(name) > 128 {
		return c.Status(fiber.StatusBadRequest).JSON(fiberNameTooLong128)
	}

	uuid, err := uuid.Parse(id)

	if err != nil {
		// return error if id is invalid
		if strings.HasPrefix(err.Error(), "invalid UUID") {
			return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidDestinationID)
		}

		log.Println("Error in parsing uuid:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	destination := db.UpdateDestinationParams{
		ID: pgtype.UUID{
			Bytes: uuid,
			Valid: true,
		},
		Column2: name,
		Column3: description,
		Column4: attraction,
		Column5: picUrl,
	}

	err = r.Queries.UpdateDestination(r.Ctx, destination)

	if err != nil {
		log.Println("Error in updatin destination in UpdateDestination db function:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "destination has been updated"})
}

// deleteDestination removes a destination from the database by ID
func (r *Repo) deleteDestination(c *fiber.Ctx) error {
	// get id
	uuid, err := uuid.Parse(c.Params("id"))

	if err != nil {
		if strings.HasPrefix(err.Error(), "invalid UUID") {
			return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidID)
		}

		log.Println("Error in parsing uuid:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	id := pgtype.UUID{
		Bytes: uuid,
		Valid: true,
	}

	err = r.Queries.DeleteDestination(r.Ctx, id)

	if err != nil {
		switch err.Error() {
		case "no rows in result set":
			return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidID)
		case "ERROR: update or delete on table \"destination\" violates foreign key constraint \"trip_destination_id_fkey\" on table \"trip\" (SQLSTATE 23503)":
			return c.Status(fiber.StatusBadRequest).JSON(
				&fiber.Map{"error": "trip exists for this destination"})
		}

		log.Println("Error in deleting destination in DeleteDestination db function:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "destination has been deleted"})
}
