package main

import (
	"flag"
	"log"
	"os"

	"market-core/internal/app/migrate"
)

func main() {
	dir := flag.String("dir", "migrations", "migrations directory")
	flag.Parse()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN is required")
	}

	if err := migrate.Run(dsn, *dir); err != nil {
		log.Fatal(err)
	}
	log.Println("migrations applied successfully")
}
