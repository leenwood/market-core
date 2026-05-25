package main

import (
	"context"
	"log"

	"market-core/internal/app/service"
)

func main() {
	if err := service.RunServer(context.Background()); err != nil {
		log.Fatal(err)
	}
}
