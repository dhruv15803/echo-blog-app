package main

import (
	"flag"
	"log"
	"os"

	"github.com/dhruv15803/echo-blog-app/db"
	"github.com/dhruv15803/echo-blog-app/scripts"
	"github.com/dhruv15803/echo-blog-app/storage"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	dbPostgresConnStr := os.Getenv("DB_CONN")

	db, err := db.ConnectToPostgres(dbPostgresConnStr)
	if err != nil {
		log.Fatal(err)
	}

	email := flag.String("email", "", "Admin user email")
	password := flag.String("password", "", "Admin user password")
	flag.Parse()

	log.Printf("Email :- %v\n", *email)
	log.Printf("password :- %v\n", *password)

	storage := storage.NewStorage(db)
	scripts := scripts.NewScripts(storage)

	if *email == "" || *password == "" {
		log.Fatal("ADMIN_EMAIL and ADMIN_PASSWORD must be set")
	}

	adminUser, err := scripts.CreateAdminUser(*email, *password)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Admin user created :- %v\n", adminUser)
}
