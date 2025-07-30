package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dhruv15803/echo-blog-app/cloudinary"
	"github.com/dhruv15803/echo-blog-app/db"
	"github.com/dhruv15803/echo-blog-app/handlers"
	"github.com/dhruv15803/echo-blog-app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Addr          string
	DbConnStr     string
	CloudinaryUrl string
}

func loadServerConfig() (*ServerConfig, error) {

	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	addr := ":" + os.Getenv("PORT")
	dbConnStr := os.Getenv("DB_CONN")
	cloudinaryUrl := os.Getenv("CLOUDINARY_URL")

	return &ServerConfig{
		Addr:          addr,
		DbConnStr:     dbConnStr,
		CloudinaryUrl: cloudinaryUrl,
	}, nil
}

func main() {

	cfg, err := loadServerConfig()
	if err != nil {
		log.Fatalln("failed to load server config")
	}

	dbConn, err := db.ConnectToPostgres(cfg.DbConnStr)
	if err != nil {
		log.Fatalf("failed to connect to postgres db :- %v\n", err.Error())
	}

	defer dbConn.Close()
	log.Println("connected to database!")

	cld, err := cloudinary.NewCloudinaryInstance()
	if err != nil {
		log.Fatalf("failed to load cloudinary instance :- %v\n", err.Error())
	}

	store := storage.NewStorage(dbConn)
	handler := handlers.NewHandler(store, cld)

	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Get("/health", handler.HealthCheckHandler)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handler.RegisterUserHandler)
			r.Post("/login", handler.LoginUserHandler)
			r.Put("/activate/{token}", handler.ActivateUserHandler)
			r.Post("/forgot-password", handler.ForgotPasswordHandler)
			r.Put("/password-reset/{token}", handler.ResetPasswordHandler)
			r.With(handler.AuthMiddleware).Get("/user", handler.GetAuthUser)
		})

		r.Route("/topic", func(r chi.Router) {
			r.With(handler.AuthMiddleware).With(handler.AdminMiddleware).Post("/", handler.CreateTopicHandler)
			r.With(handler.AuthMiddleware).With(handler.AdminMiddleware).Delete("/{topicId}", handler.DeleteTopicHandler)
			r.With(handler.AuthMiddleware).With(handler.AdminMiddleware).Put("/{topicId}", handler.UpdateTopicHandler)
			r.With(handler.AuthMiddleware).Get("/topics", handler.GetTopicsHandler)
		})

		r.Route("/blog", func(r chi.Router) {
			r.Use(handler.AuthMiddleware)
			r.Post("/", handler.CreateBlogHandler)
		})

		r.Route("/file", func(r chi.Router) {
			r.Post("/upload", handler.UploadFileHandler)
		})
	})

	server := http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		IdleTimeout:  time.Second * 30,
	}

	log.Printf("Starting server on port %v\n", cfg.Addr)

	if err = server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server on port %v\n", cfg.Addr)
	}
}
