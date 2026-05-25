// Package main is the entrypoint for the market-core HTTP server.
//
// @title           market-core API
// @version         1.0
// @description     Product catalog microservice: full-text search, filtering, categories, analytics, favorites.
//
// @contact.name   leenwood
// @contact.email  george200135@gmail.com
//
// @host      localhost:8080
// @BasePath  /api/v1
// @schemes   http
package main

import (
	"context"
	"log"

	_ "market-core/docs/swagger"
	"market-core/internal/app/service"
)

func main() {
	if err := service.RunServer(context.Background()); err != nil {
		log.Fatal(err)
	}
}
