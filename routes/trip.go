package routes

import (
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ItsYourAbhi/goTrip/db"
)

// getTrips retrieves all trips from the database
func (r *Repo) ListTrips(c *fiber.Ctx) error {
	trips, err := r.Queries.ListTrips(r.Ctx)

	if err != nil {
		log.Println("Error in getting trips in GetTrips db function:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	if len(trips) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(
			&fiber.Map{"error": "no trips found"})
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{"data": &trips})
}

// getTrip retrieves a single trip by ID
func (r *Repo) getTrip(c *fiber.Ctx) error {
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

	trip, err := r.Queries.GetTrip(r.Ctx, id)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidID)
		}

		log.Println("Error in getting trip in GetTrip db function:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{"data": &trip})
}

// Retrieves trips by Destination ID
func (r *Repo) getTripsByDestinationID(c *fiber.Ctx) error {
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

	trips, err := r.Queries.GetTripsByDestinationID(r.Ctx, id)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidID)
		}

		log.Println("Error in getting trip in GetTrip db function:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	if len(trips) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(
			&fiber.Map{"error": "no trips found"})
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{"data": &trips})
}

// createTrip adds a new trip to the database
func (r *Repo) createTrip(c *fiber.Ctx) error {
	// Extract trip details from form data
	id := c.FormValue("id")
	name := c.FormValue("name")
	startDateStr := c.FormValue("start_date")
	endDateStr := c.FormValue("end_date")
	destinationId := c.FormValue("destination_id")

	// Validate input
	if name == "" || startDateStr == "" || endDateStr == "" || destinationId == "" {
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

	startDate, err := time.Parse("2006-1-2", startDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidTimeFormat)
	}

	endDate, err := time.Parse("2006-1-2", endDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidTimeFormat)
	}

	// Parse destination UUID
	destinationUuid, err := uuid.Parse(destinationId)

	if err != nil {
		if strings.HasPrefix(err.Error(), "invalid UUID") {
			return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidDestinationID)
		}

		log.Println("Error in parsing uuid:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	// Prepare trip data for database insertion
	trip := db.CreateTripParams{
		ID: pgtype.UUID{
			Bytes: Uuid,
			Valid: true,
		},
		Name: name,
		StartDate: pgtype.Date{
			Time:  startDate,
			Valid: true,
		},
		EndDate: pgtype.Date{
			Time:  endDate,
			Valid: true,
		},
		DestinationID: pgtype.UUID{
			Bytes: destinationUuid,
			Valid: true,
		},
	}

	// Create trip in database
	err = r.Queries.CreateTrip(r.Ctx, trip)

	if err != nil {
		if err.Error() == "ERROR: insert or update on table \"trip\" violates foreign key constraint \"trip_destination_id_fkey\" (SQLSTATE 23503)" {
			return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidDestinationID)
		}

		log.Println("Error in creating trip in CreateTrip db function:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	return c.Status(fiber.StatusCreated).JSON(&fiber.Map{
		"message": "trip has been added"})
}

// updateTrip modifies an existing trip in the database
func (r *Repo) updateTrip(c *fiber.Ctx) error {
	// Extract trip details from form data
	id := c.FormValue("id")
	name := c.FormValue("name")
	startDate := c.FormValue("start_date")
	endDate := c.FormValue("end_date")
	destinationId := c.FormValue("destination_id")

	// Validate input
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiberUndefinedParamError)
	}

	if name == "" && startDate == "" && endDate == "" && destinationId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiberUndefinedParamError)
	}

	if len(name) > 128 {
		return c.Status(fiber.StatusBadRequest).JSON(fiberNameTooLong128)
	}

	var StartDate pgtype.Date
	var EndDate pgtype.Date

	if startDate != "" {
		date, err := time.Parse("2006-1-2", startDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidTimeFormat)
		}

		StartDate = pgtype.Date{
			Time:  date,
			Valid: true,
		}
	}

	if endDate != "" {
		date, err := time.Parse("2006-1-2", endDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidTimeFormat)
		}

		EndDate = pgtype.Date{
			Time:  date,
			Valid: true,
		}
	}

	var DestinationID pgtype.UUID

	if destinationId != "" {
		// Parse destination UUID
		destinationUuid, err := uuid.Parse(destinationId)

		if err != nil {
			if strings.HasPrefix(err.Error(), "invalid UUID") {
				return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidDestinationID)
			}

			log.Println("Error in parsing uuid:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
		}

		DestinationID = pgtype.UUID{
			Bytes: destinationUuid,
			Valid: true,
		}
	}

	// Parse trip UUID
	uuid, err := uuid.Parse(id)

	if err != nil {
		if strings.HasPrefix(err.Error(), "invalid UUID") {
			return c.Status(fiber.StatusBadRequest).JSON(
				&fiber.Map{"error": "invalid trip id"})
		}

		log.Println("Error in parsing uuid:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	// Prepare trip data for database update
	trip := db.UpdateTripParams{
		ID: pgtype.UUID{
			Bytes: uuid,
			Valid: true,
		},
		Column2: name,
		Column3: StartDate,
		Column4: EndDate,
		Column5: DestinationID,
	}

	// Update trip in database
	err = r.Queries.UpdateTrip(r.Ctx, trip)

	if err != nil {
		if err.Error() == "ERROR: insert or update on table \"trip\" violates foreign key constraint \"trip_destination_id_fkey\" (SQLSTATE 23503)" {
			return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidDestinationID)
		}

		log.Println("Error in updating trip in UpdateTrip db function:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "trip has been updated"})
}

// deleteTrip removes a trip from the database by ID
func (r *Repo) deleteTrip(c *fiber.Ctx) error {
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

	err = r.Queries.DeleteTrip(r.Ctx, id)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Status(fiber.StatusBadRequest).JSON(fiberInvalidID)
		}

		log.Println("Error in deleting trip in DeleteTrip db function:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiberUnknownError)
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "trip has been deleted"})
}
