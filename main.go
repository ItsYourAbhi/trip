package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	// "github.com/ansrivas/fiberprometheus/v2"
	"github.com/bytedance/sonic"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/gofiber/storage/memory"
	// "github.com/gofiber/storage/postgres/v3"
	// "github.com/gofiber/storage/redis/v3"
	// "github.com/gofiber/storage/memcache"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/ItsYourAbhi/goTrip/db"
	"github.com/ItsYourAbhi/goTrip/routes"
)

func init() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("error loading .env file:", err)
	}
}

func main() {
	// Set SSL mode, default to "disable" if not specified
	sslmode := os.Getenv("SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		// Construct database connection string
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s sslmode=%s",
			os.Getenv("HOST"), os.Getenv("USER"),
			os.Getenv("PASS"), os.Getenv("NAME"), sslmode)

		// Add port to connection string if specified
		port := os.Getenv("PORT")
		if port != "" {
			dsn += " port=" + port
		}
	}

	// Create database connection
	ctx := context.Background()
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("error connecting db pool:", err)
	}
	defer conn.Close()

	// Initialize database queries and repository
	queries := db.New(conn)
	repo := &routes.Repo{
		Ctx:     ctx,
		Queries: queries,
	}

	store := memory.New()

	// custom JSON encoder/decoder for performance
	fiberConfig := fiber.Config{
		// Prefork:     true,
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	}

	// Initializing fiber app
	app := fiber.New(fiberConfig)

	// Configure CSRF middleware
	// cookieSecure = false if err happens
	cookieSecure, _ := strconv.ParseBool(os.Getenv("CookieSecure"))
	cookieSameSite := os.Getenv("CookieSameSite")
	if cookieSameSite == "" {
		cookieSameSite = "Lax"
	}

	csrf := csrf.New(csrf.Config{
		KeyLookup:      "form:csrf",
		CookieName:     "csrf",
		ContextKey:     "csrf",
		CookieSameSite: cookieSameSite,
		CookieSecure:   cookieSecure,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Println("CSRF Error:", err)
			return c.Status(fiber.StatusForbidden).JSON(
				&fiber.Map{"error": "forbidden"})
		},
		// Storage: postgres.New(postgres.Config{
		// 	DB:    conn,
		// 	Table: "csrf_token",
		// }),
		Storage: store,
	})

	// Configure Swagger
	swagger := swagger.New(swagger.Config{
		Title:    "Trip API",
		FilePath: "swagger.yaml",
	})

	// Rate Limiter Config
	limiter := limiter.New(limiter.Config{
		Max:        3,
		Expiration: time.Second,
		Storage:    store,
	})

	// Cache Config
	cache := cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Path() == "/csrf" || c.Path() == "/csrf/"
		},
		Storage:      store,
		Expiration:   1 * time.Minute,
		CacheControl: true,
	})

	// Cors Config
	origin := os.Getenv("ALLOWED_ORIGIN")
	if origin == "" {
		origin = "http://localhost:8080, http://127.0.0.1:8080"
	}

	cors := cors.New(cors.Config{
		AllowOrigins:     origin,
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Cookie",
	})

	// prometheus config
	// prometheus := fiberprometheus.New("trip")
	// prometheus.RegisterAt(app, "/metrics")
	// prometheus.SetSkipPaths([]string{"/ping", "/favicon.ico"})

	// Middlewares: logger, swagger, recover, cache, rate limiter & CSRF protection
	app.Use(logger.New(), limiter, cors, csrf, cache, swagger, recover.New())

	// Set up routes
	repo.SetupRoutes(app)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "3000"
	}

	// Start the server
	// fmt.Println("Starting go server at port", port)
	port = ":" + port
	log.Fatal(app.Listen(port))
}
