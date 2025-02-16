package routes

import (
	"context"
	"os"

	"github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"

	"github.com/ItsYourAbhi/goTrip/db"
)

type CheckIfOwner bool

type Repo struct {
	Ctx     context.Context
	Queries *db.Queries
}

var (
	secret    []byte
	ownerUuid string

	// Defining Errors
	fiberUnknownError        = &fiber.Map{"error": "some unknown error occured"}
	fiberUndefinedParamError = &fiber.Map{"error": "some params are undefined"}
	// fiberUnauthorizedError    = &fiber.Map{"error": "unauthorized"}
	fiberInvalidID            = &fiber.Map{"error": "invalid id"}
	fiberInvalidDestinationID = &fiber.Map{"error": "invalid destination id"}
	fiberInvalidTimeFormat    = &fiber.Map{"error": "invalid time format"}
	fiberInvalidEmailPass     = &fiber.Map{"error": "invalid email or password"}
	fiberValTooLong33         = &fiber.Map{"error": "value should not be more than 33 characters"}
	fiberNameTooLong128       = &fiber.Map{"error": "name should not be more than 128 characters"}
)

func (r *Repo) SetupRoutes(app *fiber.App) {
	// loads neccessary environment variables
	loadEnvVars()

	// Prometheus
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	// Auth
	app.Get("/csrf", getCsrfToken)
	login := app.Group("/login")
	login.Post("", r.login)
	app.Post("/register", r.register)

	// For testing csrf
	app.Post("", testRoute)

	// initializing /destination route
	destination := app.Group("/destination")
	destination.Get("", r.ListDestinations)
	destination.Get("/:id", r.getDestination)

	// initializing /trip route
	trip := app.Group("/trip")
	trip.Get("", r.ListTrips)
	trip.Get("/:id", r.getTrip)
	trip.Get("/destination/:id", r.getTripsByDestinationID)

	// JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: secret},
	}))

	// JWT Routes below

	// /user route
	usr := app.Group("/user")
	// usr.Get("", aboutUser) // GET isn't protected by CSRF
	usr.Post("", aboutUser)
	usr.Put("", r.updateUser)

	// only owner can promote user to admin
	// or demote admin to user with additional admin=demote in form
	// app.Get("admin/:email", r.promoteAdmin) // GET isn't protected by CSRF
	checkOwner := CheckIfOwner(true)
	app.Post("/admin", checkOwner.checkIfAdmin, r.promoteAdmin)

	// checks for admin instead of owner
	checkOwner = false
	// Middleware to check if user is admin
	app.Use(checkOwner.checkIfAdmin)
	// test if middleware is working properly
	app.Get("", testRoute)

	// changes destination
	destination.Post("", r.createDestination)
	destination.Put("", r.updateDestination)
	destination.Delete("/:id", r.deleteDestination)

	// changes trip
	trip.Post("", r.createTrip)
	trip.Put("", r.updateTrip)
	trip.Delete("/:id", r.deleteTrip)
}

func testRoute(c *fiber.Ctx) error {
	return c.JSON(&fiber.Map{"msg": "working"})
}

// loads neccessary environment variables
func loadEnvVars() {
	// Get the secret signing Key for JWT
	secret = []byte(os.Getenv("SECRET"))
	// Get uuid of owner
	ownerUuid = os.Getenv("OWNER_UUID")

	// log.Println(secret)
	// log.Println(ownerUuid)
}
